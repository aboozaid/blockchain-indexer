package polygon

import (
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/config"
	"nodes-indexer/modules/evm"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
)

type PolygonModule interface {
	common.LifecycleModule
	GetPolygonService() PolygonService
}

type module struct{
	service PolygonService
}

func NewPolygonModule() PolygonModule {
	logger := log.Logger.With().Str("module", "PolygonModule").Logger()
	cfg := config.NewConfigModule().GetConfigService()
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
	chain := cfg.EvmChains[0]
	client, err := ethclient.Dial(chain.RPCUrl)
	if err != nil {
		panic(err.Error())
	}
	// pool, err := ants.NewPool(3) // create 3 threads per blockchain
	// if err != nil {
	// 	panic(err.Error())
	// }
	evmModule := evm.NewEvmModule()
	service := NewPolygonService(/*pool,*/ client, chain.ChainID, evmModule.GetEvmService(), &logger)

	logger.Info().Str("Module", "PolygonModule").Msg("Polygon module initialized successfully")

	return module{service}
}

func (m module) GetPolygonService() PolygonService {
	return m.service
}

func (m module) OnAppStart() error {
	return m.service.OnModuleStart()
}

func (m module) OnAppDestroy() error {
	return m.service.OnModuleStop()
}