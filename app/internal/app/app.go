package app

import (
	"context"
	"shortner/internal/config"
	"shortner/internal/database"
	"shortner/internal/handler"
	"shortner/internal/repository"
	"shortner/internal/service"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

type URLShortenerApp struct {
	cfg *config.Config

	engine *ginext.Engine

	log *zlog.Zerolog
}

func NewURLShortenerApp(cfg *config.Config, log *zlog.Zerolog) (*URLShortenerApp, error) {
	r := ginext.New("release")

	strategy := retry.Strategy{
		Attempts: cfg.Retry.Attempts,
		Delay:    cfg.Retry.Delay,
		Backoff:  cfg.Retry.Backoff,
	}

	opts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}

	db, err := database.Connect(cfg.DB.URL, []string{}, opts)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to connect to database")
		return nil, err
	}

	urlsRepo := repository.NewURLRepository(db, strategy)
	clicksRepo := repository.NewClicksRepository(db, strategy)

	shortnererService := service.NewURLShortnererService(urlsRepo, clicksRepo)

	shortenerHandler := handler.NewURLHandler(shortnererService, log)

	r.GET("/", func(c *ginext.Context) {
		c.File("public/index.html")
	})

	r.Engine.Use(ginext.Logger())
	r.Engine.Use(ginext.Recovery())

	shortenerHandler.RegisterRoutes(r)

	return &URLShortenerApp{
		cfg:    cfg,
		engine: r,
		log:    log,
	}, nil
}

func (a *URLShortenerApp) Run(ctx context.Context) {
	if err := a.engine.Run(":" + a.cfg.App.Port); err != nil {
		a.log.Error().
			Err(err).
			Msg("failed to run server")
		return
	}
}
