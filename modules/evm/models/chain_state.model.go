package models

type ChainState struct {
	ChainID         int64 `sql:"primary_key"`
	LastBlockNumber *string
	LastBlockHash   *string
	CreatedAt       int64
	UpdatedAt       int64
}
