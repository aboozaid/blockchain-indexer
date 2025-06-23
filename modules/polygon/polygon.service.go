package polygon

import (
	"fmt"
	"nodes-indexer/modules/common"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog"
)

type PolygonService interface {
	common.LifecycleService
}

type ChainOption struct {
	ID      int64
	Client  *ethclient.Client
	BlocksConfirmations int8
}

type service struct {
	pool *ants.Pool
	client  *ethclient.Client
	wg sync.WaitGroup

	logger *zerolog.Logger

	latest_checked_block_number uint64
	latest_checked_block_hash string
}

func NewPolygonService(pool *ants.Pool, client *ethclient.Client /*chainOptions []ChainOption,*/, logger *zerolog.Logger) PolygonService {
	// chains := map[int64]*chain{}
	// for _, chainOption := range chainOptions {
	// 	chains[chainOption.ID] = &chain{
	// 		ChainOption:               chainOption,
	// 		next_block_number:         0,
	// 		latest_checked_block_hash: ethcommon.Hash{},
	// 	}
	// }
	return &service{
		pool:       pool,
		client: client,
		logger:     logger,
	}
}

func (s *service) trackAndCollectBlocks() {
	
}

func (s *service) OnModuleStart() error {
	
	s.pool.Submit(func() {
		s.trackAndCollectBlocks()
	})

	return nil
}

func (s *service) OnModuleStop() error {
	defer s.pool.Release()

	// for _, chain := range s.chains {
	// 	chain.stopIndexing()
	// }
	// s.wg.Wait()
	
	fmt.Println("Polygon service stopped")
	return nil
}