package main

import (
	"nodes-indexer/modules"

	"github.com/rs/zerolog/log"
)

func main() {	
	app := modules.NewAppModule()
	cfg := app.GetConfig()
	
	app.OnStart(func() {
		log.Info().Msgf("Server is running on localhost:%d", cfg.App.Port)
	})
}