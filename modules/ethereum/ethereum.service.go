package ethereum

import "nodes-indexer/modules/evm"

type EthereumService interface {
	StartIndexing()
}

type service struct {
	evmService evm.EvmService
	// client *ethclient.Client
	// next_block_number uint64
	// latest_checked_block_hash common.Hash
}

// const (
// 	BlocksThreshold = 14
// 	PauseTime = 20 * time.Second // 3 pauses per minute
// )

// var ExtraOne = big.NewInt(1)

func NewEthereumService(evmService evm.EvmService) EthereumService {
	return service{evmService}
}

/*
Must be called once the app started
*/
func (s service) StartIndexing() {

	// We may have read that from database and if nil read from the blockchain
	// last_block_no, err := s.client.BlockNumber(context.Background())
	// if err != nil {
	// 	panic(err)
	// }
	// s.latest_block_no = &last_block_no
	// for {
	// 	last_block_number, err := s.client.BlockNumber(context.Background())
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	if s.next_block_number == 0 {
	// 		s.next_block_number = last_block_number
	// 	} else {
	// 		threshold := last_block_number - s.next_block_number
	// 		if threshold >= BlocksThreshold {
	// 			from_block := new(big.Int).SetUint64(s.next_block_number)
	// 			headers := make([]*types.Header, BlocksThreshold)
	// 			// headers[0] = s.latest_block
	// 			// headers[BlocksThreshold-1] = current_last_block
	// 			for i := 0; i <= BlocksThreshold - 1; i-- {
	// 				headers[i], err = s.client.HeaderByNumber(context.Background(), from_block)
	// 				if err != nil {
	// 					panic(err.Error())
	// 				}
	// 				from_block.Add(from_block, ExtraOne)
	// 			}
	// 			s.latest_checked_block_hash = headers[BlocksThreshold-1].Hash()
	// 			s.next_block_number = headers[BlocksThreshold-1].Number.Uint64()
	// 		}
	// 	}

	// 	time.Sleep(PauseTime)
	// }
	// s.evmService.StartIndexing()
}