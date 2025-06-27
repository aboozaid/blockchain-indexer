package models

type Block struct {
	ChainID         int64
	BlockNumber     string
	BlockHash       string
	BlockParentHash string
	BlockConfirmed  *int64
	CreatedAt       int64
	UpdatedAt       int64
}