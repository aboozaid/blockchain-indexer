package tron

import (
	"fmt"
	"nodes-indexer/modules/common"
)

type TronModule interface {
	common.LifecycleModule
	// GetEvmService() EvmService
}

type module struct{
	// service EvmService
}

func NewTronModule() TronModule {
	// cfg := config.NewConfigModule().GetConfigService()
	// chainOptions := make([]ChainOption, 0, len(cfg.EvmChains))
	// for _, chain := range cfg.EvmChains {
	// 	client, err := ethclient.Dial(chain.RPCUrl)
	// 	if err != nil {
	// 		panic(err.Error())
	// 	}
				
	// 	chainOptions = append(chainOptions, ChainOption{
	// 		ID:      chain.ChainID,
	// 		Client:  client,
	// 		BlocksConfirmations: chain.Confirmations,
	// 	})
	// }
	// service := NewEvmService(chainOptions)
	
	return module{}
}

// func (m module) GetEvmService() EvmService {
// 	return m.service
// }

func (m module) OnAppStart() error {
	// call service
	fmt.Println("TronModule initialized")
	return nil
}

func (m module) OnAppDestroy() error {
	// call service
	fmt.Println("TronModule is being destroyed")
	fmt.Println("TronModule destroyed")
	return nil
}