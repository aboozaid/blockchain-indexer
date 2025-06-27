package polygon

import (
	"context"
	"fmt"
	"maps"
	"math/big"
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/evm"
	"nodes-indexer/modules/evm/dto"
	"nodes-indexer/modules/evm/models"
	polDto "nodes-indexer/modules/polygon/dto"
	"slices"
	"strconv"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
)

type PolygonService interface {
	common.LifecycleService
}

type transaction struct {
	block_no	uint64
	tx_hash 	string
	from 	string
	to	string
	value	*hexutil.Big
	smart_contract_address	string
}

type service struct {
	pool *ants.Pool
	client  *ethclient.Client
	evmService evm.EvmService
	stopIndexing context.CancelFunc 
	wg sync.WaitGroup

	logger *zerolog.Logger

	chainID int64
	// latest_checked_block_number uint64
	latest_checked_block_hash string
}

const (
	COLLECT_BLOCKS_RANGE = 6
	COLLECT_BLOCK_BY_BLOCK_INTERVAL = 100 * time.Millisecond // 10 blocks per second
	TRACK_AND_COLLECT_BLOCKS_INTERVAL = 20 * time.Second // 3 iteration per minute

)

var (
	INC_BY_ONE = big.NewInt(1)
	TRANSFER_TOPIC_HASH = ethcommon.HexToHash("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
)

func NewPolygonService(/*pool *ants.Pool,*/ client *ethclient.Client, chainID string, evmService evm.EvmService, logger *zerolog.Logger) PolygonService {
	// chains := map[int64]*chain{}
	// for _, chainOption := range chainOptions {
	// 	chains[chainOption.ID] = &chain{
	// 		ChainOption:               chainOption,
	// 		next_block_number:         0,
	// 		latest_checked_block_hash: ethcommon.Hash{},
	// 	}
	// }
	id, _ := strconv.ParseInt(chainID, 10, 64)

	return &service{
		//pool:       pool,
		client: client,
		evmService: evmService,
		logger:     logger,
		chainID: id,
	}
}

func (s *service) trackAndCollectBlocks(ctx context.Context, reorge_channel chan []*types.Header) error {
	state, err := s.evmService.GetChainState(ctx, s.chainID)
	if err != nil {
		return err
	}
	var next_block_number uint64
	if state.LastBlockNumber != nil && state.LastBlockHash != nil {
		last_block_number, _ := strconv.ParseUint(*state.LastBlockNumber, 10, 64)
		next_block_number = last_block_number+1
		s.latest_checked_block_hash = *state.LastBlockHash
		s.logger.Info().Msgf("Start tracking and collecting from block: %d", last_block_number)
	}
	
	for {
		select {
			case <- ctx.Done():
				return s.saveState(context.Background(), next_block_number-1)
			default:
				/* NOTE:  1 * 3 iteration/minute = rpc calls in minute */
				latest_chain_block_number, err := s.client.BlockNumber(ctx)
				if err != nil {
					s.logger.Error().Err(err)
					return err
				}
				if next_block_number == 0 {
					next_block_number = latest_chain_block_number
				} else {
					blocks_range := latest_chain_block_number - next_block_number
					/* TODO: We need dynamic block range based on how far range between blockchain and the indexer */
					if blocks_range >= COLLECT_BLOCKS_RANGE {
						s.logger.Debug().Msgf("Start indexing %d blocks", COLLECT_BLOCKS_RANGE)
						blocks_headers := make([]*types.Header, COLLECT_BLOCKS_RANGE)
						from_block := new(big.Int).SetUint64(next_block_number)
						/* NOTE: COLLECT_BLOCKS_RANGE * 3 iteration/minute = rpc calls in minute */
						for i := range COLLECT_BLOCKS_RANGE {
							/* FIXME: When context cancelled during loop it will not break */
							// select {
							// case <- ctx.Done():
							// 	return s.saveState(context.Background(), next_block_number-1)
							// default:
							// 	h, err := s.client.HeaderByNumber(ctx, from_block)
							// 	if err != nil {
							// 		s.logger.Error().Err(err)
							// 		return err
							// 	}
							// 	blocks_headers[i] = h
							// 	from_block.Add(from_block, INC_BY_ONE)
							// 	time.Sleep(COLLECT_BLOCK_BY_BLOCK_INTERVAL)
							// }
							h, err := s.client.HeaderByNumber(ctx, from_block)
							if err != nil {
								s.logger.Error().Err(err)
								if ctx.Err() != nil {
									fmt.Println("Context cancelled")
								}
								return err
							}
							blocks_headers[i] = h
							from_block.Add(from_block, INC_BY_ONE)
							time.Sleep(COLLECT_BLOCK_BY_BLOCK_INTERVAL)
						}
						/* REVIEW: When to set next block after process blocks or before */
						next_block_number = blocks_headers[COLLECT_BLOCKS_RANGE-1].Number.Uint64()+1

						// if cap is full it will be queued until it gets free
						reorge_channel <- blocks_headers
					}
				}
				s.logger.Debug().Msg("Sleep until Fetch next blocks")
				time.Sleep(TRACK_AND_COLLECT_BLOCKS_INTERVAL)
		}
	}
}

func (s *service) saveState(ctx context.Context, latest_block_number uint64) error {
	s.logger.Debug().Msg("Saving latest chain state to the database")
	err := s.evmService.UpdateChainState(context.Background(), s.chainID, &dto.UpdateChainStateDto{
		LastBlockNumber: strconv.FormatUint(latest_block_number, 10),
		LastBlockHash: s.latest_checked_block_hash,
	})
	return err
 }



func (s *service) handleReorg(ctx context.Context, reorge_channel chan []*types.Header, blocks_channel chan []*types.Header) error {
	for headers := range reorge_channel {
		from_block := headers[0].Number
		to_block := headers[COLLECT_BLOCKS_RANGE-1].Number
		s.logger.Debug().Msgf("Checking reorg from block %d to block %d", from_block, to_block)

		if is_forked, forked_block_index := s.evmService.IsReorgDetected(headers, ethcommon.HexToHash(s.latest_checked_block_hash)); is_forked {
			s.logger.Info().Msgf("Reorg found at block %d and start backtracking over the chain", headers[*forked_block_index].Number)
		}
	
		s.logger.Debug().Msgf("No reorg found from block %d to block %d", from_block, to_block)
		
		blocks_channel <- headers
	}
	return nil
}

func (s *service) traceBlockTransactions(ctx context.Context, block *types.Header) ([]polDto.Transaction, error) {
	var block_traces []polDto.BlockTrace
	err := s.client.Client().CallContext(ctx, &block_traces, "trace_block", block.Number)
	if err != nil {
		return []polDto.Transaction{}, err
	}

	block_txs := make([]polDto.Transaction, 0, len(block_traces)/2)
	for _, trace := range block_traces {
		if trace.Error != nil || trace.Type != "call" || trace.Action.Value == "" {
			continue
		}
		value, err := hexutil.DecodeBig(trace.Action.Value)
		if err != nil {
			continue
		}
		// check value greater than 0
		if value.Sign() > 0 && (trace.Action.CallType == "call" || trace.Action.CallType == "delegatecall") {
			// native coin transfer
			transaction := polDto.Transaction{
				From: trace.Action.From,
				To: trace.Action.To,
				Hash: trace.TransactionHash,
				BlockNumber: block.Number.Uint64(),
				Value: trace.Action.Value,
				Timestamp: block.Time,
			}
			block_txs = append(block_txs, transaction)
		}

		if trace.IsTokenTransfer() {
			transaction := polDto.Transaction{
				Hash: trace.TransactionHash,
				BlockNumber: block.Number.Uint64(),
				Timestamp: block.Time,
			}

			extracted := transaction.ExtractTokenInfo(trace)
			if extracted {
				block_txs = append(block_txs, transaction)
			}
		}
	}

	return block_txs, nil
}

func (s *service) handleBlocks(ctx context.Context, blocks_channel chan []*types.Header) error {
	var wg sync.WaitGroup
	for blocks := range blocks_channel {
		trace_results_channel := make(chan []polDto.Transaction, COLLECT_BLOCKS_RANGE)
		for _, block := range blocks {
			wg.Add(1)
			// this will block if no pool available until a one get free
			err := s.runTaskOnPool(func() {
				defer wg.Done()
				// TODO: We need to handle errors
				txs, _ := s.traceBlockTransactions(ctx, block)
				// if err != nil {
				// 	panic(err)
				// }
				trace_results_channel <- txs
			})
			if err != nil {
				return err
			}
		}
		wg.Wait()
		close(trace_results_channel)

		transactions_contracts := make(map[string]struct{}) // used map to prevent duplicated addresses
		transactions_from_to_addresses := make(map[string][]uint64) // address -> []string -> block numbers
		blocks_transactions := make(map[uint64][]polDto.Transaction)
		addOrUpdateAddress := func (address string, block_number uint64)  {
			blocks, isExists := transactions_from_to_addresses[address]
			if !isExists {
				transactions_from_to_addresses[address] = []uint64{block_number}
			} else if block_number != 0 {
				exists := slices.Contains(blocks, block_number)
				if !exists {
					transactions_from_to_addresses[address] = append(blocks, block_number)
				}
			}
		}
		// collect transactions from each block
		for transactions := range trace_results_channel {
			if len(transactions) > 0 {
				blocks_transactions[transactions[0].BlockNumber] = transactions
				var block_number uint64
				for _, tx := range transactions {
					if tx.Contract != "" {
						transactions_contracts[tx.Contract] = struct{}{}
					}
					if block_number != 0 {
						block_number = 0
					}

					if tx.BlockNumber != transactions[0].BlockNumber {
						block_number = tx.BlockNumber
					}
					addOrUpdateAddress(tx.From, block_number)
					addOrUpdateAddress(tx.To, block_number)
				}
			}
		}

		addresses := make([]string, len(transactions_from_to_addresses))
		i := 0
		for k := range transactions_from_to_addresses {
			addresses[i] = k
			i++
		}
		results, err := s.evmService.SearchByAddresses(ctx, s.chainID, addresses)
		if err != nil {
			return err
		}
		if len(results) > 0 {
			// we have transfers in/our from our wallets
			contracts := slices.Collect(maps.Keys(transactions_contracts))
			supported_contracts, err := s.evmService.GetSupportedContracts(ctx, s.chainID, contracts)
			if err != nil {
				return err
			}
			for _, wallet := range results {
				wallet_address := wallet.Address
				address_blocks := transactions_from_to_addresses[wallet_address]
				for _, address_block := range address_blocks {
					transactions := blocks_transactions[address_block]
					var filtered_address_transactions []polDto.Transaction
					for _, tx := range transactions {
						if tx.From == wallet_address || tx.To == wallet_address {
							if tx.Contract != ""  {
								if supported := slices.ContainsFunc(supported_contracts, func(c *models.CryptoCurrency) bool {
									return c.ContractAddress == tx.Contract
								}); !supported {
									continue
								}
							}
							filtered_address_transactions = append(filtered_address_transactions, tx)
						}
					}
					for _, tx := range filtered_address_transactions {
						if tx.From == wallet_address {
							// withdraw
							if tx.Contract != "" {
								s.logger.Debug().Msgf("Whoa! the address: %s just transferred %d value of contract %s", wallet_address, hexutil.MustDecodeBig(tx.Value), tx.Contract)
							} else {
								s.logger.Debug().Msgf("Whoa! the address: %s just transferred %d value of native coin", wallet_address, hexutil.MustDecodeBig(tx.Value))
							}
						} else {
							// deposit
							if tx.Contract != "" {
								s.logger.Debug().Msgf("Whoa! the address: %s just received %d value of contract %s", wallet_address, hexutil.MustDecodeBig(tx.Value), tx.Contract)
							} else {
								s.logger.Debug().Msgf("Whoa! the address: %s just received %d value of native coin", wallet_address, hexutil.MustDecodeBig(tx.Value))
							}
						}
					}
				}
			}
		}
		s.evmService.CreateBlocks(ctx, s.chainID, blocks)
	}
	return nil
}

func (s *service) OnModuleStart() error {
	pool, err := ants.NewPool(COLLECT_BLOCKS_RANGE/2 + 3)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	
	s.pool = pool
	s.stopIndexing = cancel

	// submitTask := func(task func()) error {
	// 	err := s.pool.Submit(func() {
	// 		// defer s.wg.Done()
	// 		task()
	// 	})
	// 	if err != nil {
	// 		cancel()
	// 		return err
	// 	}
	// 	return nil
	// }
	reorge_channel := make(chan []*types.Header, 3) // cap of 3 as we iterate 3 times over the minute
	// blocks_channel := make(chan []*types.Header, 1)

	s.wg.Add(1) // add two tasks

	err = s.runTaskOnPool(func() { 
		defer s.wg.Done()
		if err := s.trackAndCollectBlocks(ctx, reorge_channel); err != nil {
			s.logger.Error().Err(err)
			return
		}
	}) 
	if err != nil {
		s.logger.Error().Err(err)
		return err
	}

	// if err := s.runTaskOnPool(func() { 
	// 	defer s.wg.Done()
	// 	s.handleReorg(ctx, reorge_channel, blocks_channel) 
	// }); err != nil {
	// 	return err
	// }

	// if err := s.runTaskOnPool(func() { 
	// 	defer s.wg.Done()
	// 	s.handleBlocks(ctx, blocks_channel)
	// }); err != nil {
	// 	return err
	// }

	return nil
}

func (s *service) OnModuleStop() error {
	s.logger.Debug().Msg("Stopping Service")

	defer s.pool.Release()

	// for _, chain := range s.chains {
	// 	chain.stopIndexing()
	// }
	s.stopIndexing()
	s.logger.Debug().Msg("Cancelling context")
	s.wg.Wait()
	
	return nil
}

func (s *service) runTaskOnPool(task func()) error {
	if err := s.pool.Submit(task); err != nil {
		return err
	}
	return nil
}

// func (s *service) decodeValueHex(val string) string {

// 	if len(val) < 2 || val == "0x0" {
// 		return "0"
// 	}

// 	if val[:2] == "0x" {
// 		x, err := DecodeBig(val)

// 		if err != nil {
// 			// log.Error("errorDecodeValueHex", "str", val, "err", err)
// 		}
// 		return x.String()
// 	} else {
// 		x, ok := big.NewInt(0).SetString(val, 16)

// 		if !ok {
// 			// log.Error("errorDecodeValueHex", "str", val, "ok", ok)
// 		}

// 		return x.String()
// 	}
// }
