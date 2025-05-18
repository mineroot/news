package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"github.com/mineroot/news/config"
	"github.com/mineroot/news/internal/db"
	"github.com/mineroot/news/internal/handlers"
	"github.com/mineroot/news/internal/posts"
	"github.com/mineroot/news/internal/route"
)

func main() {
	// load config
	cfg, err := config.LoadConfig()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// setup logger
	var out io.Writer = os.Stderr
	if !cfg.IsProd() {
		out = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.DateTime, NoColor: cfg.IsProd()}
	}
	logger := log.Output(out).With().Caller().Logger().Level(cfg.LogLevel())
	logger.Info().
		Stringer("config", cfg).
		Send()
	ctx := logger.WithContext(context.Background())

	// respect os signals
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// run app
	if err := run(ctx, cfg); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("unexpected exit")
		os.Exit(1)
	}
	logger.Debug().Msg("successfully exited")
}

func run(ctx context.Context, cfg *config.Config) error {
	logger := zerolog.Ctx(ctx)

	logger.Debug().Msg("connecting to mongodb")
	mongoClient, err := db.CreateMongoClient(ctx, cfg.MongoUri())
	if err != nil {
		return err
	}
	logger.Debug().Msg("mongodb is up")
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

	postRepo := posts.NewRepository(db.GetPostsCollection(mongoClient))
	validation := validator.New(validator.WithRequiredStructEnabled())

	e := echo.New()
	e.HTTPErrorHandler = handlers.ErrorHandler(logger)
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		DisableStackAll:   true,
		DisablePrintStack: true,
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			uri, _ := url.PathUnescape(c.Request().RequestURI)
			logger.Error().
				Str("URI", uri).
				Err(err).
				Msg("server panic [recovered]")
			return err
		},
	}))

	// for docker healthcheck
	e.HEAD("/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	e.GET(route.ViewHome, handlers.ViewHomeHandler(postRepo)).Name = route.ViewHome

	e.GET(route.ViewPost, handlers.ViewPostHandler(postRepo)).Name = route.ViewPost

	e.GET(route.ViewCreatePostForm, handlers.ViewCreatePostFormHandler()).Name = route.ViewCreatePostForm
	e.POST(route.CreatePost, handlers.CreatePostHandler(postRepo, validation)).Name = route.CreatePost

	e.GET(route.ViewUpdatePostForm, handlers.ViewUpdatePostFormHandler(postRepo)).Name = route.ViewUpdatePostForm
	e.POST(route.UpdatePost, handlers.UpdatePostHandler(postRepo, validation)).Name = route.UpdatePost

	e.POST(route.DeletePost, handlers.DeletePostHandler(postRepo)).Name = route.DeletePost

	e.GET(route.ViewSearch, handlers.ViewSearchHandler(postRepo)).Name = route.ViewSearch

	g, ctx := errgroup.WithContext(ctx)
	// run http server
	g.Go(func() error {
		addr := fmt.Sprintf(":%s", cfg.HttpServerPort())
		// ignore http.ErrServerClosed as it's expected after e.Shutdown() call
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})
	// gracefully shutdown http server after ctx cancellation
	g.Go(func() error {
		<-ctx.Done()
		shoutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = e.Shutdown(shoutdownCtx)
		return nil
	})

	return g.Wait()
}
