//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package tables

import (
	"github.com/go-jet/jet/v2/sqlite"
)

var ChainsStates = newChainsStatesTable("", "chains_states", "ChainState")

type chainsStatesTable struct {
	sqlite.Table

	// Columns
	ChainID         sqlite.ColumnInteger
	LastBlockNumber sqlite.ColumnString
	LastBlockHash   sqlite.ColumnString
	CreatedAt       sqlite.ColumnInteger
	UpdatedAt       sqlite.ColumnInteger

	AllColumns     sqlite.ColumnList
	MutableColumns sqlite.ColumnList
	DefaultColumns sqlite.ColumnList
}

type ChainsStatesTable struct {
	chainsStatesTable

	EXCLUDED chainsStatesTable
}

// AS creates new ChainsStatesTable with assigned alias
func (a ChainsStatesTable) AS(alias string) *ChainsStatesTable {
	return newChainsStatesTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new ChainsStatesTable with assigned schema name
func (a ChainsStatesTable) FromSchema(schemaName string) *ChainsStatesTable {
	return newChainsStatesTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new ChainsStatesTable with assigned table prefix
func (a ChainsStatesTable) WithPrefix(prefix string) *ChainsStatesTable {
	return newChainsStatesTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new ChainsStatesTable with assigned table suffix
func (a ChainsStatesTable) WithSuffix(suffix string) *ChainsStatesTable {
	return newChainsStatesTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newChainsStatesTable(schemaName, tableName, alias string) *ChainsStatesTable {
	return &ChainsStatesTable{
		chainsStatesTable: newChainsStatesTableImpl(schemaName, tableName, alias),
		EXCLUDED:          newChainsStatesTableImpl("", "excluded", ""),
	}
}

func newChainsStatesTableImpl(schemaName, tableName, alias string) chainsStatesTable {
	var (
		ChainIDColumn         = sqlite.IntegerColumn("chain_id")
		LastBlockNumberColumn = sqlite.StringColumn("last_block_number")
		LastBlockHashColumn   = sqlite.StringColumn("last_block_hash")
		CreatedAtColumn       = sqlite.IntegerColumn("created_at")
		UpdatedAtColumn       = sqlite.IntegerColumn("updated_at")
		allColumns            = sqlite.ColumnList{ChainIDColumn, LastBlockNumberColumn, LastBlockHashColumn, CreatedAtColumn, UpdatedAtColumn}
		mutableColumns        = sqlite.ColumnList{LastBlockNumberColumn, LastBlockHashColumn, CreatedAtColumn, UpdatedAtColumn}
		defaultColumns        = sqlite.ColumnList{CreatedAtColumn, UpdatedAtColumn}
	)

	return chainsStatesTable{
		Table: sqlite.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ChainID:         ChainIDColumn,
		LastBlockNumber: LastBlockNumberColumn,
		LastBlockHash:   LastBlockHashColumn,
		CreatedAt:       CreatedAtColumn,
		UpdatedAt:       UpdatedAtColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
		DefaultColumns: defaultColumns,
	}
}
