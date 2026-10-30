package server

import (
	"context"
	"fmt"
	"net/http"
	"rcon-scheduler-panel/internal/config"
	"rcon-scheduler-panel/internal/domain/auth"
	"rcon-scheduler-panel/internal/domain/job"
	"rcon-scheduler-panel/internal/domain/user"
	"rcon-scheduler-panel/internal/middleware"
	"rcon-scheduler-panel/internal/server/database"
	"rcon-scheduler-panel/internal/validation"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"gorm.io/gorm"
)

type App interface {
	Start()
	Shutdown()
}

type app struct {
	DB     *gorm.DB
	Gin    *gin.Engine
	Config *config.Config
	Srv    *http.Server
}

func NewApp(cfg *config.Config) (App, error) {
	db, err := database.Setup(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to start database!")
		return nil, err
	}

	validation.Register()
	// s := NewScheduler(cfg)

	r := gin.New()
	r.Use(middleware.Logger())
	r.Use(middleware.ErrorHandler())
	r.Use(gin.CustomRecovery(middleware.Recovery()))
	r.Use(cors.Default())

	r.HandleMethodNotAllowed = true
	r.NoRoute(middleware.NoRoute())
	r.NoMethod(middleware.NoMethod())

	api := r.Group("/api")

	userRepo := user.NewRepository(db)
	jobRepo := job.NewRepository(db)
	userSVC := user.NewService(userRepo)
	authSVC := auth.NewService(userSVC, cfg)
	jobSVC := job.NewService(jobRepo)
	
	MakeRoutes(api, 
		auth.NewHandler(authSVC), 
		user.NewHandler(userSVC, cfg), 
		job.NewHandler(jobSVC, cfg),
	)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.AppPort),
		Handler:           r.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &app{
		DB:     db,
		Config: cfg,
		Gin:    r,
		Srv:    srv,
	}, nil
}

func (a *app) Start() {
	go func() {
		log.Info().Msg("Starting application")
		if err := a.Srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("listen failed")
		}
	}()
}

func (a *app) Shutdown() {
	log.Info().Msg("Shutting down application")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Error shutting down")
		return
	}
	log.Info().Msg("Server gracefully stopped")
}
