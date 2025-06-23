package models

type CryptoCurrency struct {
	ChainID         int64
	ContractAddress string
	Symbol          string
	Decimals        int64
	CreatedAt       int64
	UpdatedAt       int64
}
