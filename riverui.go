package riverui

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/riverui"
	"github.com/riverqueue/riverui/ui"
	"github.com/rs/cors"
)

type Handler struct {
	logger         *slog.Logger
	riveruiHandler http.Handler
	dbPool         *pgxpool.Pool
}

func init() {
	caddy.RegisterModule(Handler{})
	httpcaddyfile.RegisterHandlerDirective("riverui", parseRiverUIHandlerDirective)
}

// CaddyModule returns the Caddy module information.
func (Handler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.riverui",
		New: func() caddy.Module { return new(Handler) },
	}
}

//go:embed dist/*
var index embed.FS

// Provision sets up the River UI handler.
func (h *Handler) Provision(ctx caddy.Context) error {
	h.logger = ctx.Slogger()

	// TODO: ~gc/vite-server replacement of variables at runtime
	// TODO: after replacing, recalculate hashes, if needed?

	// set the default River UI web app to embedded build
	ui.Index = index

	dbURL := os.Getenv("DATABASE_URL")   // TODO: make configurable through Caddyfile too
	dbPool, err := getDBPool(ctx, dbURL) // TODO: make lazy during provisioning
	if err != nil {
		return fmt.Errorf("error connecting to db: %w", err)
	}
	h.dbPool = dbPool

	corsOrigins := []string{"*"} // TODO: fix; make configurable
	corsHandler := cors.New(cors.Options{
		AllowedMethods: []string{"GET", "HEAD", "POST", "PUT"},
		AllowedOrigins: corsOrigins,
	})

	client, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})
	if err != nil {
		return fmt.Errorf("error creating river client: %w", err)
	}

	//logger := slog.Default() // TODO: support RIVER_DEBUG; log level
	pathPrefix := "/"

	handlerOpts := &riverui.HandlerOpts{
		Client: client,
		DBPool: dbPool,
		Logger: h.logger,
		Prefix: pathPrefix,
	}

	handler, err := riverui.NewHandler(handlerOpts)
	if err != nil {
		return fmt.Errorf("error creating handler: %w", err)
	}

	// TODO: wrap logging, otel, metrics; similar to the riverui binary?
	h.riveruiHandler = corsHandler.Handler(handler)

	return nil
}

// Validate ensures the handler's configuration is valid.
func (h *Handler) Validate() error {
	// TODO: implement validation
	return nil
}

// ServeHTTP is the Caddy handler for serving HTTP requests
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	h.riveruiHandler.ServeHTTP(w, r)
	return nil
}

func (h *Handler) Cleanup() error {
	h.dbPool.Close()
	return nil
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (h *Handler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	// TODO: parse additional handler directives (none exist now)
	return nil
}

// parseRiverUIHandlerDirective parses the `riverui` Caddyfile directive
func parseRiverUIHandlerDirective(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	var handler Handler
	err := handler.UnmarshalCaddyfile(h.Dispenser)
	return handler, err
}

func getDBPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing db config: %w", err)
	}

	dbPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to db: %w", err)
	}
	return dbPool, nil
}

// Interface guards
var (
	_ caddy.Module                = (*Handler)(nil)
	_ caddy.Provisioner           = (*Handler)(nil)
	_ caddy.Validator             = (*Handler)(nil)
	_ caddy.CleanerUpper          = (*Handler)(nil)
	_ caddyhttp.MiddlewareHandler = (*Handler)(nil)
	_ caddyfile.Unmarshaler       = (*Handler)(nil)
)
