package main

import (
	"context"
	"os"
	"os/signal"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/logger"
	"rcon-scheduler-panel/internal/server"
	"syscall"

	"github.com/rs/zerolog/log"
)

//	@title			RCON Scheduler Panel API
//	@version		1.0
//	@description	API for RCON Scheduler Panel.

//	@BasePath	/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter "Bearer <token>" (token from /auth/login).

func main() {
	logger.Setup()
	log.Info().Msg("Starting server")

	cfg, err := config.NewConfig(".")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}
	logger.UpdateLogLevel(cfg.LogLevel)

	app, err := server.NewApp(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to init app")
	}
	app.Start()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Debug().Msg("Waiting for shutdown signal")
	<-ctx.Done()

	app.Shutdown()
}
