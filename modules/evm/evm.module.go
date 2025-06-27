package evm

import (
	"nodes-indexer/modules/database"

	"github.com/rs/zerolog/log"
)

type EvmModule interface {
	// common.LifecycleModule
	GetEvmService() EvmService
}

type module struct{
	service EvmService
}

func NewEvmModule() EvmModule {
	logger := log.Logger.With().Str("module", "EvmModule").Logger()
	// cfg := config.NewConfigModule().GetConfigService()
	// chainOptions := make([]ChainOption, 0, len(cfg.EvmChains))
	// for _, chain := range cfg.EvmChains {
	// 	client, err := ethclient.Dial(chain.RPCUrl)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
	// 	id, _ := strconv.ParseInt(chain.ChainID, 10, 64)	
	// 	chainOptions = append(chainOptions, ChainOption{
	// 		ID:     id,
	// 		Client:  client,
	// 		BlocksConfirmations: chain.Confirmations,
	// 	})
	// }
	// pool, err := ants.NewPool(len(cfg.EvmChains) * 3) // create 3 threads per blockchain
	// if err != nil {
	// 	panic(err.Error())
	// }
	service := NewEvmService(/*pool, chainOptions,*/ NewEvmRespository(database.NewDatabaseModule().GetDB()), &logger)
	
	return module{service}
}

func (m module) GetEvmService() EvmService {
	return m.service
}

// func (m module) OnAppStart() error {
// 	return m.service.OnModuleStart()
// }

// func (m module) OnAppDestroy() error {
// 	return m.service.OnModuleStop()
// }