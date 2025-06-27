package evm

import (
	"context"
	"database/sql"
	"nodes-indexer/jet/tables"
	"nodes-indexer/modules/database"
	"nodes-indexer/modules/evm/models"

	"github.com/go-jet/jet/v2/qrm"
	jet "github.com/go-jet/jet/v2/sqlite"
)

type EvmRepository interface {
	// GetBlocks(context.Context, string, int) (map[string]*models.Block, error)
	GetConfirmedReorgedBlocks(context.Context, int64, []string) ([]*models.Block, error)
	GetLatestChainState(context.Context, int64) (*models.ChainState, error)
	GetBlockByHash(context.Context, int64, string) (*models.Block, error)
	GetWalletAddresses(context.Context, int64, []string) ([]*models.Address, error)
	CreateBlocks(context.Context, int64, func (*[]*models.Block) error) error
	GetSupportedCryptoCurrencies(context.Context, int64, []string) ([]*models.CryptoCurrency, error)
	UpdateChainState(context.Context, int64, func (*models.ChainState) error) error
}

type repository struct{
	db *sql.DB
}

func NewEvmRespository(db *sql.DB) EvmRepository {
	return repository{db}
}

func (r repository) GetConfirmedReorgedBlocks(ctx context.Context, chainID int64, blocks_numbers []string) ([]*models.Block, error) {
	var jet_blocks_numbers []jet.Expression
	for _, block_number := range blocks_numbers {
		jet_blocks_numbers = append(jet_blocks_numbers, jet.String(block_number))
	}

	var rows []*models.Block
	tbl := tables.Blocks
	stmt := tbl.SELECT(tbl.BlockHash, tbl.BlockNumber, tbl.ChainID, tbl.CreatedAt).
		FROM(tbl).
		WHERE(
			tbl.ChainID.EQ(jet.Int64(chainID)).
			AND(tbl.BlockConfirmed.EQ(jet.Int32(1))).
			AND(tbl.BlockNumber.IN(jet_blocks_numbers...)),
			// AND(
			// 	tbl.BlockNumber.GT(jet.Int64(fromBlock)),
			// ).
			// AND(
			// 	tbl.BlockNumber.LT_EQ(jet.Int64(toBlock)),
			// ),
		)
		// ORDER_BY(tbl.CreatedAt.ASC())

	err := stmt.QueryContext(ctx, database.GetDB(ctx, r.db), &rows)
	
	if err != nil {
		return nil, err
	}
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return rows, nil
}

func (r repository) GetLatestChainState(ctx context.Context, chainID int64) (*models.ChainState, error) {
	var row models.ChainState

	tbl := tables.ChainsStates
	stmt := tbl.SELECT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	FROM(tbl).
	WHERE(tbl.ChainID.EQ(jet.Int64(chainID)))

	err := stmt.QueryContext(ctx, database.GetDB(ctx, r.db), &row)
	if err != nil && err == qrm.ErrNoRows {
		return &models.ChainState{}, nil
	}
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return &row, err
}

func (r repository) GetBlockByHash(ctx context.Context, chainID int64, hash string) (*models.Block, error) {
	var row models.Block

	tbl := tables.Blocks
	stmt := tbl.SELECT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	FROM(tbl).
	WHERE(tbl.ChainID.EQ(jet.Int64(chainID)).AND(tbl.BlockHash.EQ(jet.String(hash))))

	err := stmt.QueryContext(ctx, database.GetDB(ctx, r.db), &row)
	if err != nil && err == qrm.ErrNoRows {
		return nil, nil
	}
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return &row, err
}

func (r repository) CreateBlocks(ctx context.Context, chainID int64, f func (*[]*models.Block) error) (error) {
	blocks := new([]*models.Block)
	err := f(blocks)
	if err != nil{
		return err
	}
	tbl := tables.Blocks
	stmt := tbl.INSERT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	MODELS(blocks)

	_, err = stmt.ExecContext(ctx, database.GetDB(ctx, r.db))
	// if err != nil {
	// 	return err
	// }
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return err
}

func (r repository) GetWalletAddresses(ctx context.Context, chainID int64, addresses []string) ([]*models.Address, error) {
	var jet_addresses []jet.Expression
	for _, address := range addresses {
		jet_addresses = append(jet_addresses, jet.String(address))
	}

	var rows []*models.Address

	tbl := tables.Addresses
	stmt := tbl.SELECT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	FROM(tbl).
	WHERE(tbl.ChainID.EQ(jet.Int64(chainID)).AND(tbl.Address.IN(jet_addresses...)))

	err := stmt.QueryContext(ctx, database.GetDB(ctx, r.db), &rows)

	if err != nil {
		return nil, err
	}
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return rows, nil
}

func (r repository) GetSupportedCryptoCurrencies(ctx context.Context, chainID int64, contracts []string) ([]*models.CryptoCurrency, error) {
	var jet_contracts []jet.Expression
	for _, contract := range contracts {
		jet_contracts = append(jet_contracts, jet.String(contract))
	}

	var rows []*models.CryptoCurrency

	tbl := tables.CryptoCurrencies
	stmt := tbl.SELECT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	FROM(tbl).
	WHERE(tbl.ChainID.EQ(jet.Int64(chainID)).AND(tbl.ContractAddress.IN(jet_contracts...)))

	err := stmt.QueryContext(ctx, database.GetDB(ctx, r.db), &rows)

	if err != nil {
		return nil, err
	}
	// if e, ok := err.(*pgconn.PgError); ok {
	// 	if e.Code == "23505" {
	// 		return errors.ErrRowExists
	// 	}
	// }
	return rows, nil
}

func (r repository) UpdateChainState(ctx context.Context, chainID int64, f func (*models.ChainState) error) error {
	chain_state := new(models.ChainState)
	err := f(chain_state)
	if err != nil {
		return err
	}

	tbl := tables.ChainsStates
	stmt := tbl.INSERT(
		tbl.AllColumns.Except(
			tbl.CreatedAt,
			tbl.UpdatedAt,
		),
	).
	MODEL(chain_state).
	ON_CONFLICT(tbl.ChainID).
	WHERE(tbl.ChainID.EQ(jet.Int64(chainID))).
	DO_UPDATE(
		jet.SET(
			tbl.ChainID.SET(jet.Int64(chain_state.ChainID)),
			tbl.LastBlockHash.SET(jet.String(*chain_state.LastBlockHash)),
			tbl.LastBlockNumber.SET(jet.String(*chain_state.LastBlockNumber)),
		),
	)
	_, err = stmt.ExecContext(ctx, database.GetDB(ctx, r.db))

	return err
}