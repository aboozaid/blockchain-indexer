package evm

import (
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/config"
	"nodes-indexer/modules/database"
	"strconv"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
	"github.com/rs/zerolog/log"
)

type EvmModule interface {
	common.LifecycleModule
	GetEvmService() EvmService
}

type module struct{
	service EvmService
}

func NewEvmModule() EvmModule {
	logger := log.Logger.With().Str("module", "EvmModule").Logger()
	cfg := config.NewConfigModule().GetConfigService()
	chainOptions := make([]ChainOption, 0, len(cfg.EvmChains))
	for _, chain := range cfg.EvmChains {
		client, err := ethclient.Dial(chain.RPCUrl)
		if err != nil {
			panic(err.Error())
		}
		id, _ := strconv.ParseInt(chain.ChainID, 10, 64)
		pool, err := ants.NewPool(int(chain.BatchBlocksRange+3)) // create 3 threads per blockchain
		if err != nil {
			panic(err.Error())
		}
		logger := log.Logger.With().Str("chain", chain.Name).Logger()
		chainOptions = append(chainOptions, ChainOption{
			ID:     id,
			Client:  client,
			BlockConfirmations: chain.Confirmations,
			Pool: pool,
			Logger: &logger,
		})
	}
	logger.Info().Msg("Evm module initialized successfully")

	service := NewEvmService(/*pool,*/ chainOptions, NewEvmRespository(database.NewDatabaseModule().GetDB()))
	
	return module{service}
}

func (m module) GetEvmService() EvmService {
	return m.service
}

func (m module) OnAppStart() error {
	return m.service.OnModuleStart()
}

func (m module) OnAppDestroy() error {
	return m.service.OnModuleStop()
}