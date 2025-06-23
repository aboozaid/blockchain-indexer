package evm

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/evm/models"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
)

type EvmService interface {
	common.LifecycleService
}

type ChainOption struct {
	ID      int64
	Client  *ethclient.Client
	BlocksConfirmations int8
}

type chain struct {
	ChainOption

	stopIndexing context.CancelFunc

	next_block_number uint64
	latest_checked_block_hash ethcommon.Hash
}

type transaction struct {
	hash string
	contract_address string
}

type service struct {
	pool *ants.Pool
	chains map[int64]*chain
	wg sync.WaitGroup

	repository EvmRepository
	logger *zerolog.Logger
}


var (
	next_block = big.NewInt(1)
	BlocksThreshold = 6 // Number of blocks to batch process
)

const (
	BLOCKS_BATCH_SLEEP_INTERVAL = 10 * time.Second // used when we still did not reach threshold
	BLOCKS_BATCH_SLEEP_INTERVAL_PER_THRESHOLD = 8 * time.Second // used when we reach threshold
	BLOCK_BY_BLOCK_SLEEP_INTERVAL = 100 * time.Millisecond // 10 blocks per second
	REORG_TRAVERSE_NODE_TIMEOUT = 16 * time.Minute // 16 minute including BLOCK_BY_BLOCK_SLEEP_INTERVAL so it actually 8 min without sleeping
)

func NewEvmService(pool *ants.Pool, chainOptions []ChainOption, repository EvmRepository, logger *zerolog.Logger) EvmService {
	chains := map[int64]*chain{}
	for _, chainOption := range chainOptions {
		chains[chainOption.ID] = &chain{
			ChainOption:         chainOption,
			next_block_number:  0,
			latest_checked_block_hash: ethcommon.Hash{},
		}
	}
	return &service{
		pool: pool, 
		chains: chains,
		repository: repository,
		logger: logger,
	}
		
}

func (s *service) OnModuleStart() error {
	for chainID, chain := range s.chains {
		// NOTE: Needs to be updated according to how many goroutines running below
		s.wg.Add(2)

		ctx, cancel := context.WithCancel(context.Background())
		
		submitTask := func(task func()) error {
			err := s.pool.Submit(func() {
				defer s.wg.Done()
				task()
			})
			if err != nil {
				cancel()
				return err
			}
			return nil
		}
		receiveBlocks := make(chan []*types.Block, 3) // Buffered channel to receive blocks, 3 because we batch 3 times per minute
		processBlocks := make(chan []*types.Block, 1) // Buffered channel to process blocks
		// TODO: Handle errors if returned
		if err := submitTask(func() { s.trackAndCollectBlocks(ctx, chainID, receiveBlocks) }); err != nil {
			return err
		}
		if err := submitTask(func() { s.handleReOrg(ctx, chainID, receiveBlocks, processBlocks) }); err != nil {
			return err
		}
		// if err := submitTask(func() { s.handleBlocks(ctx, chainID) }); err != nil {
		// 	return err
		// }

		chain.stopIndexing = cancel
	}
	return nil
}

func (s *service) OnModuleStop() error {
	defer s.pool.Release()

	for _, chain := range s.chains {
		chain.stopIndexing()
	}
	s.wg.Wait()
	
	fmt.Println("EVM service stopped")
	return nil
}
// called by multiple chain goroutine
// each chain have one trackAndCollectBlocks single goroutine
// trackAndCollectBlocks called by multiple goroutines on different chains
func (s *service) trackAndCollectBlocks(ctx context.Context, chainID int64, receiveBlocks chan []*types.Block) error {
	chain_state, err := s.repository.GetLatestChainState(ctx, chainID) // -> called by multiple goroutine at the same time
	if err != nil {
		return err
	}
	
	chain := s.chains[chainID] //NOTE: -> read from map
				
	if chain_state.LastBlockNumber != nil && chain_state.LastBlockHash != nil {
		last_checked_block_number, _ := strconv.ParseUint(*chain_state.LastBlockNumber, 10, 64)
		// NOTE: At this moment only one goroutine per chain will access this variable so mutex not needed here
		chain.next_block_number = last_checked_block_number+1	//NOTE: -> write to chain ptr
		chain.latest_checked_block_hash = ethcommon.HexToHash(*chain_state.LastBlockHash) //NOTE: -> write to chain ptr
	}

	for {
		select {
			case <- ctx.Done():
				defer close(receiveBlocks)
				fmt.Println("Terminating TrackAndCollectBlocks for chain", chainID)
				chain := s.chains[chainID]
				err := s.repository.UpdateChainState(context.Background(), chainID, func(chain_state *models.ChainState) error {
					last_block_number := strconv.FormatUint(chain.next_block_number, 10)
					last_block_hash := chain.latest_checked_block_hash.String()
					chain_state.ChainID = chainID
					chain_state.LastBlockNumber = &last_block_number
					chain_state.LastBlockHash = &last_block_hash

					return nil
				})
				if err != nil {
					return err
				}
				return nil
			default:
				last_block_number, err := chain.Client.BlockNumber(ctx)
				if err != nil {
					return err
				}
				if chain.next_block_number == 0 /* NOTE: read from chain ptr */ {
					chain.next_block_number = last_block_number	/* NOTE: write to chain ptr */
					time.Sleep(BLOCKS_BATCH_SLEEP_INTERVAL)
					continue
				}
				threshold := last_block_number - chain.next_block_number	/* NOTE: read from chain ptr */
				fmt.Println("threshold: ", threshold)
				if int(threshold) >= BlocksThreshold {
					from_block := new(big.Int).SetUint64(chain.next_block_number /* NOTE: read from chain ptr */)
					blocks := make([]*types.Block, BlocksThreshold)
					for i := range BlocksThreshold {
						block, err := chain.Client.BlockByNumber(ctx, from_block) /* NOTE: read from chain ptr */
						if err != nil {
							return err
						}
						blocks[i] = block
						from_block.Add(from_block, next_block)

						// We need to sleep to not exceed Node rate limit per seconds
						time.Sleep(BLOCKS_BATCH_SLEEP_INTERVAL_PER_THRESHOLD)
					}
					//TODO:  latest_checked_block_hash should be updated after checking re-org not here
					// chain.latest_checked_block_hash = headers[BlocksThreshold-1].Hash()
					chain.next_block_number = blocks[BlocksThreshold-1].NumberU64()+1	/* NOTE: write to chain ptr */

					select {
						case receiveBlocks <- blocks:
							fmt.Println("Sent blocks for re-org handling on", chainID, ":", len(blocks))
						default:
							fmt.Println("Receive channel is full, skipping sending blocks for chain", chainID)
					}
					
					time.Sleep(BLOCKS_BATCH_SLEEP_INTERVAL_PER_THRESHOLD)
					continue
				}
				time.Sleep(BLOCKS_BATCH_SLEEP_INTERVAL)
		}
	}
}

func (s *service) handleReOrg(ctx context.Context, chainID int64, receiveBlocks chan []*types.Block, processBlocks chan []*types.Block) error {
	chain := s.chains[chainID]
	// close channel will end this loop as well
	for blocks := range receiveBlocks {
		// select {
		// 	case <- ctx.Done():
		// 		fmt.Println("Terminating HandleReOrg for chain", chainID)
		// 		return nil
		// 	case blocks := <- receiveBlocks:
		// 		fmt.Println("Received blocks for re-org handling on", chainID, ":", len(blocks), "from block:", blocks[0].Number, "to block:", blocks[len(blocks)-1].Number)
		// }
		fmt.Println("Received blocks for re-org handling on", chainID, ":", len(blocks), "from block:", blocks[0].Number, "to block:", blocks[len(blocks)-1].Number)

		forked, forked_block_index := s.isReorgDetected(chain.latest_checked_block_hash, blocks)
		if forked {
			// we have to backtrack the node to get the ancestor block
			total_blocks := len(blocks)
			new_chain_forked_block := blocks[*forked_block_index]
			new_chain_ancestor_block_index := -1
			// var old_chain_latest_block *types.Block
			// if *forked_block_index-1 > 0 {
			// 	old_chain_latest_block = blocks[*forked_block_index-1]
			// }
			// new_blocks := []*types.Block{}
			
			// var new_chain_ancestor_block *types.Block
			// var new_chain_ancestor_block_index int

			/* slice blocks for only old blocks + can we remove new_blocks? */
			// blocks_by_hashes := make(map[string]int, len(blocks))
			// for index, b := range blocks[:*forked_block_index] {
			// 	hash := b.Hash().Bytes()
			// 	blocks_by_hashes[string(hash)] = index
			// }

			new_chain_previous_block_hash := new_chain_forked_block.ParentHash()

			for i := *forked_block_index-1; i>=0; i-- {
				new_chain_previous_block, err := chain.Client.BlockByHash(ctx, new_chain_previous_block_hash)
				if err != nil {
					return err
				}
				if blocks[i].Hash() == new_chain_previous_block.Hash() {
					// find ancestor block
					// new_chain_ancestor_block = blocks[i]
					new_chain_ancestor_block_index = i
					break
				}
				// if index, found := blocks_by_hashes[string(new_chain_previous_block.Hash().Bytes())]; found {
				// 	// find ancestor block
				// 	new_chain_ancestor_block = blocks[index]
				// 	break
				// }
				// NOTE: Append the new block on the array so that we process it later
				// new_blocks = append(new_blocks, new_chain_previous_block)
				blocks[i] = new_chain_previous_block
				if i != 0 {
					new_chain_previous_block_hash = new_chain_previous_block.ParentHash() // get new chain previous block and continue
					time.Sleep(BLOCK_BY_BLOCK_SLEEP_INTERVAL)
				}
			}

			// if ancestor block nil we need to check database either
			if new_chain_ancestor_block_index == -1 {
				// total_new_blocks := len(new_blocks)
				// if *forked_block_index > 0 {
				// 	new_chain_previous_block_hash = /*new_blocks[len(new_blocks)-1].ParentHash()*/ blocks[0].ParentHash()
				// } else {
				// 	// new_blocks = append(new_blocks, blocks[0])
				// }
				new_chain_previous_block_hash = blocks[0].ParentHash()
				// NOTE: We need timer in order to end the loop in case of we could not find the ancestor
				ctx, cancel := context.WithTimeout(ctx, REORG_TRAVERSE_NODE_TIMEOUT)
				should_stop := false
				defer cancel()
				for !should_stop {
					select {
						case <- ctx.Done():
							should_stop = true
						default:
							new_chain_previous_block, err := chain.Client.BlockByHash(ctx, new_chain_previous_block_hash)
							if err != nil {
									return err
							}
							// NOTE: Index and cache should be used here to optimize the performance and RAM usage
							db_block, err := s.repository.GetBlockByHash(ctx, chainID, new_chain_previous_block.Hash().Hex())
							if err != nil {
								return err
							}
							if db_block != nil {
								// new_chain_ancestor_block = new_chain_previous_block
								cancel()
							}
							// new_blocks = append(new_blocks, new_chain_previous_block)
							blocks = append([]*types.Block{new_chain_previous_block}, blocks...)

							new_chain_previous_block_hash = new_chain_previous_block.ParentHash()
							time.Sleep(BLOCK_BY_BLOCK_SLEEP_INTERVAL) // 100 ms sleep > 10 sleep per second > 600 sleep per minute > 4800 sleep on 8 min
					}
				}
			}

			// NOTE: This case should not happen and if it did we must terminate the thread
			if new_chain_ancestor_block_index == -1 {
				return errors.New("unexpected error occured and could not find re-organization ancestor block")
			}
			
			orphaned_blocks_numbers := []string{}
			high := (len(blocks) - total_blocks) + *forked_block_index
			low := new_chain_ancestor_block_index
			// exclude ancestor block from fetched by the loop
			if low > 0 {
				low = low-1
			}
			for _, b := range blocks[low:high] {
				orphaned_blocks_numbers = append(orphaned_blocks_numbers, b.Number().String())
			}
			// if *forked_block_index > 0 {
			// 	toBlock := blocks[*forked_block_index-1].NumberU64()
			// } else {
			// 	block, err := s.repository.GetBlockByHash(ctx, chainID, chain.latest_checked_block_hash.Hex())
			// 	if err != nil {
			// 		return err
			// 	}
			// 	toBlock := block.BlockNumber
			// }

			_, err := s.repository.GetConfirmedReorgedBlocks(ctx, chainID, orphaned_blocks_numbers)
			if err != nil {
				return err
			}
			/* 
				notify if one of the orphaned_blocks confirmed before via webhook
				stop any confirmation gorouting for any orphaned block
				send new blocks to be process again
			*/

			// we now have to check if we need to notify any affected client via webhooks if that ancestor block was already confirmed and became orphaned
			// the rule of thumb is any block after ancestor_block should also be dropped
			// we need to also check if any confirmation goroutines are running and have an affected block so we have to stop it as it became orphaned

		} else {
			// no-reorg detected
			select {
				case processBlocks <- blocks:
					chain.latest_checked_block_hash = blocks[len(blocks)-1].Hash()	/* NOTE: write to chain ptr */
					fmt.Println("Sent blocks for processing on", chainID, ":", len(blocks))
				default:
					fmt.Println("Process channel is full, skipping sending blocks for chain", chainID)
			}
		}
	}
	return nil
}

func (s *service) handleBlocks(ctx context.Context, chainID int64, processBlocks chan []*types.Block) error {
	for blocks := range processBlocks {
		// select {
		// 	case <- ctx.Done():
		// 		fmt.Println("Terminating HandleBlocks for chain", chainID)
		// 		return nil
		// 	default:
		// 		fmt.Println("Handling blocks on", chainID, "...")
		// 		time.Sleep(5 * time.Second) // Simulate block handling interval
		// }
		fmt.Println("Received blocks for processing handling on", chainID, ":", len(blocks), "from block:", blocks[0].Number, "to block:", blocks[len(blocks)-1].Number)
			
		// loop on blocks and extract wallet addresses to/from
		// check database if any address matches
		// get transaction receipt of matched address
		// check transaction success and if it smart_contract or native
		// if smart_contract we check if we support that token or not
		// run another goroutine to wait for number of confirmations
		// save blocks to database
		
		// address > []string{tx_id}
		addresses_transactions := map[string][]transaction{}
		appendTxToAddress := func (address string, tx_hash string, contract_address string)  {
			if address != "" {
				addresses_transactions[address] = append(addresses_transactions[address], transaction{tx_hash, contract_address})
			}
		}
		for _, block := range blocks {
			transactions := block.Transactions()
			for _, tx := range transactions {
				if tx.To() == nil {
					continue
				}
				// WETH Transfer(address indexed src, address indexed dst, uint wad);
				// ERC20 transfer(address _to, uint256 _value)

				// two different way to listen for erc20/weth and native coins

				/* 
					trace_block 
					debug_traceBlockByNumber
					trace_transaction
					debug_traceTransaction
				
				*/
				tx_hash := tx.Hash().Hex()
				data := tx.Data()
				var from_address, to_address, contract_address string
				switch {
					case len(data) == 0:
						// native transaction ex. eth/bnb/pol etc
						sender, err := s.getTxSender(tx)
						if err != nil {
							fmt.Printf("unable to get sender of tx %s: %v\n", tx_hash, err)
							continue
						}
						from_address = sender.Hex()
						to := tx.To()
						if to != nil {
							to_address = to.Hex()
						}
					case len(data) >= 4 && strings.HasPrefix(ethcommon.Bytes2Hex(data[:4]), "a9059cbb") && len(data) > 68:
						sender, err := s.getTxSender(tx)
						if err != nil {
							fmt.Printf("unable to get sender of tx %s: %v\n", tx_hash, err)
							continue
						}
						to := ethcommon.HexToAddress("0x" + ethcommon.Bytes2Hex(data[16:36]))
						from_address = sender.Hex()
						to_address = to.Hex()
						contract_address = tx.To().Hex()
				}
				appendTxToAddress(from_address, tx_hash, contract_address)
				appendTxToAddress(to_address, tx_hash, contract_address)
			}
		}
		keys := make([]string, 0, len(addresses_transactions))
		i := 0
		for k := range addresses_transactions {
			keys[i] = k
			i++
		}
		addresses, err := s.repository.GetWalletAddresses(ctx, chainID, keys)
		if err != nil {
			return err
		}

		if len(addresses) > 0 {
			contract_addresses := []string{}
			for _, address := range addresses {
				txs := addresses_transactions[address.Address]
				for _, tx := range txs {
					if tx.contract_address != "" {
						contract_addresses = append(contract_addresses, tx.contract_address)
					}
				}
			}
			crypto_currencies, err := s.repository.GetSupportedCryptoCurrencies(ctx, chainID, contract_addresses)
			if err != nil {
				return err
			}
			chain := s.chains[chainID]
			for _, address := range addresses {
				// we need first to check contract addresses
				txs := addresses_transactions[address.Address]
				for _, tx := range txs {
					is_contract_supported := tx.contract_address != "" && slices.ContainsFunc(crypto_currencies, func(currency *models.CryptoCurrency) bool {
						return currency.ContractAddress == tx.contract_address
					})
					if is_contract_supported || tx.contract_address == "" {
						tx_receipt, err := chain.Client.TransactionReceipt(ctx, ethcommon.HexToHash(tx.hash))
						if err != nil {
							return err
						}
						is_tx_success := tx_receipt.Status
						if is_tx_success == 1 {
							// TODO: send webhook event
						}
					}
				}
			}
		}
	}
	return nil
}
func (s *service) triggerWebhook() {}

// func (s service) StartIndexing() {
// 	ctx := context.Background()
// 	for {
// 		last_block_number, err := s.client.BlockNumber(ctx)
// 		if err != nil {
// 			panic(err.Error())
// 		}
// 		if s.next_block_number == 0 {
// 			s.next_block_number = last_block_number
// 		} else {
// 			threshold := last_block_number - s.next_block_number
// 			if int(threshold) >= s.config.BlocksBatchThreshold {
// 				from_block := new(big.Int).SetUint64(s.next_block_number)
// 				headers := make([]*types.Header, s.config.BlocksBatchThreshold)
// 				// headers[0] = s.latest_block
// 				// headers[BlocksThreshold-1] = current_last_block
// 				for i := 0; i <= s.config.BlocksBatchThreshold - 1; i-- {
// 					headers[i], err = s.client.HeaderByNumber(ctx, from_block)
// 					if err != nil {
// 						panic(err.Error())
// 					}
// 					from_block.Add(from_block, ExtraOne)
// 				}
// 				s.latest_checked_block_hash = headers[s.config.BlocksBatchThreshold-1].Hash()
// 				s.next_block_number = headers[s.config.BlocksBatchThreshold-1].Number.Uint64()
// 			}
// 		}

// 		time.Sleep(s.config.BlocksBatchSleepInterval)
// 	}
// }

/* 
	Check blocks re-org and if so return state, new forked block index, ancestor block index if found
*/
func (s *service) isReorgDetected(latest_checked_block_hash ethcommon.Hash, blocks []*types.Block) (bool, *int) {
	// for this case we have to search on database to find ancestor block
	if latest_checked_block_hash != blocks[0].ParentHash() {
		current := 0
		return true, &current
	}

	for current := 1; current < len(blocks); current++ {
		// previous := current - 1
		if blocks[current].ParentHash() != blocks[current - 1].Hash() {
			//ANCHOR - Not needed because we have a new chain now we have to backtrack the node
			// for i := previous; i>=0; i-- {
			// 	if blocks[current].ParentHash == blocks[i].Hash() {
			// 		return true, &current, &i
			// 	}
			// }
			return true, &current
		}
	}
	return false, nil
}

func (s *service) getTxSender(tx *types.Transaction) (ethcommon.Address, error) {
	return types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
}