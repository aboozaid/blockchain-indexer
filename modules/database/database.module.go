package database

import (
	"database/sql"
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/config"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

// We need to define database module as globally
var instance *module

type DatabaseModule interface {
	GetDatabaseService() DatabaseService
	GetDB() *sql.DB
	common.LifecycleModule
}

type module struct {
	db *sql.DB
	service DatabaseService
}

func NewDatabaseModule() DatabaseModule {
	if instance == nil {
		cfg := config.NewConfigModule().GetConfigService()
		db, err := sql.Open("sqlite3", cfg.DB.DSN)
		
		if err != nil {
			panic(err.Error())
		}

		db.SetMaxOpenConns(max(4, runtime.NumCPU()))

		log.Info().Str("Module", "DatabaseModule").Msg("Database module initialized successfully")

		dbModule := module{
			db,
			NewDatabaseService(db, &sql.TxOptions{}),
		}
		instance = &dbModule
	}

	return instance
}

func (m module) GetDB() *sql.DB {
	return m.db
}

func (m module) OnAppStart() error {
	return nil
}

func (m module) OnAppDestroy() error {
	err := m.db.Close()
	if err != nil {
		return err
	}
	log.Debug().Str("Module", "DatabaseModule").Msg("Database module destroyed")
	return nil
}

func (m module) GetDatabaseService() DatabaseService {
	return m.service
}