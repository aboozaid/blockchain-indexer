package dto

import (
	"github.com/aboozaid/validation"
	"github.com/aboozaid/validation/is"
)

type AppConfig struct {
	Prefix string	`env:"PREFIX" envDefault:"api/v1"`
	Port   int16	`env:"PORT" envDefault:"3030"`
}

func (c AppConfig) Validate() error {
	return validation.ValidateStruct(&c, validation.Field(&c.Prefix, validation.Length(1, 10)), validation.Field(&c.Port, validation.Required))
}

type DBConfig struct {
	// Host     string	`env:"HOST"`
	// Port     int16	`env:"PORT" envDefault:"5432"`
	// User string	`env:"USER"`
	// Password string	`env:"PASSWORD"`
	DSN string	`env:"DSN"`
	Name	string	`env:"NAME"`
}

func (c DBConfig) Validate() error {
	return validation.ValidateStruct(&c,  validation.Field(&c.DSN, validation.Required), validation.Field(&c.Name, validation.Required, validation.Length(1, 80)))
}

type ChainConfig struct {
	RPCUrl     string	`env:"RPC_URL"`
	ChainID     string	`env:"CHAIN_ID"`
	Chain     string	`env:"CHAIN"`
	Confirmations int8	`env:"CONFIRMATIONS" envDefault:"12"`
}

func (c ChainConfig) Validate() error {
	return validation.ValidateStruct(&c, 
		validation.Field(&c.RPCUrl, validation.Required, is.URL), 
		validation.Field(&c.Chain, validation.Required, validation.In(
			"eth",
			"bsc",
			"btc",
			"tron",
			"solana",
			"base",
			"polygon",
			"arbitrum",
		)), 
		validation.Field(&c.ChainID, validation.Required, validation.In(
			// Mainnets
			"1",      // Ethereum Mainnet
			"56",     // Binance Smart Chain (BSC) Mainnet
			"137",    // Polygon Mainnet
			"8453",	  // Base Mainnet
			"42161",  // Arbitrum One Mainnet
			"0",      // Bitcoin (not EVM, placeholder)
			"101",    // Solana (not EVM, placeholder)
			"11111",  // Tron (not EVM, placeholder)
			// Testnets
			"5",      // Ethereum Goerli
			"11155111", // Ethereum Sepolia
			"97",     // BSC Testnet
			"80001",  // Polygon Mumbai
			"421613", // Arbitrum Goerli
			"421614", // Arbitrum Sepolia
			"111",    // Solana Testnet (placeholder)
			"2",      // Bitcoin Testnet (placeholder)
			"12345",  // Tron Testnet (placeholder)
		)),
		validation.Field(&c.Confirmations, validation.Min(3), validation.Max(50)),
	)
}

type Config struct {
	App AppConfig	`envPrefix:"APP_"`
	DB  DBConfig	`envPrefix:"DB_"`
	EvmChains []ChainConfig	`envPrefix:"EVM_"`
}

func (c Config) Validate() error {
	return validation.ValidateStruct(&c, 
		validation.Field(&c.App), 
		validation.Field(&c.DB),
		validation.Field(&c.EvmChains),
	)
}