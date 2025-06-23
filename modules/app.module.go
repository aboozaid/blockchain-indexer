package modules

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"nodes-indexer/modules/common"
	"nodes-indexer/modules/config"
	"nodes-indexer/modules/config/dto"
	"nodes-indexer/modules/database"
	"nodes-indexer/modules/evm"
	"nodes-indexer/modules/tron"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type AppModule interface {
	GetConfig() *dto.Config
	// OnInit() error
	OnStart(func())
	// OnDestroy(context.Context/*destroyed chan bool*/) error
}

type module struct {
	server *echo.Echo
	// pool   *ants.Pool
	config *dto.Config
	modules []common.LifecycleModule
}

// const MAX_POOL_SIZE = 10 // increment 1 for every new chain

func NewAppModule() AppModule {
	
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	e := echo.New()
	e.HideBanner = true

	// pool, err := ants.NewPool(MAX_POOL_SIZE)
	// if err != nil {
	// 	panic(err.Error())
	// }

	cfgModule := config.NewConfigModule()
	dbModule := database.NewDatabaseModule()

	evmModule := evm.NewEvmModule()
	tronModule := tron.NewTronModule()


	modulesLifecycle := []common.LifecycleModule{
		evmModule,
		tronModule,
		dbModule,
	}
	
	return &module{e, cfgModule.GetConfigService().Config, modulesLifecycle}
}

func (m *module) GetConfig() *dto.Config {
	return m.config
}

func (m *module) OnStart(onStarted func()) {
	// defer m.pool.Release()

	for _, mod := range m.modules {
		if err := mod.OnAppStart(); err != nil {
			log.Error().Err(err).Msg("Failed to initialize the application")
			return
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	
	serverErr := make(chan error, 1)
	serverReady := make(chan struct{}, 1)

	go func() {
		host := fmt.Sprintf("127.0.0.1:%d", m.config.App.Port)
		l, err := net.Listen("tcp", host)
		if err != nil {
			serverErr <- err
		}
		
		close(serverReady)

		s := &http.Server{
			Addr: host,
			Handler: m.server,
		}
		
		if err := s.Serve(l); err != http.ErrServerClosed {
			// e.Logger.Fatalf("Shutting down the server: %v", err)
			serverErr <- err
		}
	}()

	select {
		case <- serverReady:
			
			onStarted()

			<- ctx.Done()
			log.Info().Msg("Shutdown signal received, shutting down gracefully...")
			// <-ctx.Done()
		
			// // destroyed := make(chan bool, 1)
			// if err := app.OnDestroy(); err != nil {
			// 	// e.Logger.Fatalf("Failed to destroy the application: %v", err)
			// 	log.Fatal().Err(err).Msg("Failed to destroy the application")
			// }
			// <-destroyed
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := m.onDestroy(ctx); err != nil {
				// e.Logger.Fatalf("Unable to shutdown the server: %v", err)
				log.Fatal().Err(err).Msg("Unable to shutdown the server")
			}
		
		case err := <-serverErr:
			log.Error().Err(err).Msg("Server encountered an error and will shut down")
	}
}

func (m *module) onDestroy(ctx context.Context) error {
	// defer m.pool.Release()

	for _, mod := range m.modules {
		if err := mod.OnAppDestroy(); err != nil {
			return err
		}
	}
	
	if err := m.server.Shutdown(ctx); err != nil {
		return err
	}
	// destroyed <- true
	return nil
}