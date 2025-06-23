package ethereum

import (
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/panjf2000/ants/v2"
)

type EthereumModule interface {
	OnStop()
}

type module struct{
	client *ethclient.Client
}

/* 
	ethereum service > evm service
*/

func NewEthereumModule(pool *ants.Pool) EthereumModule {
	client, err := ethclient.Dial("https://bsc-testnet.core.chainstack.com/0f298922bd1ee0c68f91f49a066be05b")
	if err != nil {
		panic(err.Error())
	}
	
	// evmModule := evm.NewEvmModule(client)
	// service := NewEthereumService(evmModule.GetEvmService())
	// pool.Submit(service.StartIndexing)
	
	return module{client}
}

func (m module) OnStop() {
	m.client.Close()
}