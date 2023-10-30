package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/Hofsiedge/person-api/internal/api"
	"github.com/Hofsiedge/person-api/internal/completer"
	"github.com/Hofsiedge/person-api/internal/config"
	"github.com/Hofsiedge/person-api/internal/repo/postgres"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gorilla/mux"
	middleware "github.com/oapi-codegen/nethttp-middleware"
)

type loggingTransport struct {
	logger *slog.Logger
}

func (s *loggingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	var bytes, respBytes []byte
	bytes, _ = httputil.DumpRequestOut(r, false)

	resp, err := http.DefaultTransport.RoundTrip(r)

	if resp != nil {
		respBytes, _ = httputil.DumpResponse(resp, true)
	}

	s.logger.Debug("completer client was used",
		slog.String("request", string(bytes)),
		slog.String("response", string(respBytes)),
	)

	return resp, err //nolint:wrapcheck
}

func makeCompleter(logger *slog.Logger) *completer.Completer {
	completerCfg, err := config.Read[config.CompleterConfig]()
	if err != nil {
		log.Fatal(err)
	}

	//nolint:exhaustruct
	completerHTTPClient := http.Client{
		Transport: &loggingTransport{
			logger: logger,
		},
		Timeout: completer.Timeout,
	}

	return completer.New(completerCfg, &completerHTTPClient)
}

//nolint:funlen
func main() {
	serverCfg, err := config.Read[config.ServerConfig]()
	if err != nil {
		log.Fatal(err)
	}

	// logger
	level := serverCfg.LogLevel
	if serverCfg.Debug {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   serverCfg.Debug,
		Level:       level,
		ReplaceAttr: nil,
	}))

	// postgres repo
	pgCfg, err := config.Read[config.PostgresConfig]()
	if err != nil {
		log.Fatal(err)
	}

	people, err := postgres.New(pgCfg)
	if err != nil {
		log.Fatal(err)
	}

	// external API client
	comp := makeCompleter(logger)

	server, err := api.New(people, comp, logger)
	if err != nil {
		log.Fatal(err)
	}

	// validator
	spec, err := api.GetSwagger()
	if err != nil {
		err = fmt.Errorf("error creating OpenAPI validator: %w", err)
		log.Fatal(err)
	}

	//nolint:exhaustruct
	spec.Servers = openapi3.Servers{&openapi3.Server{URL: "/api/v0"}}
	oapiValidator := middleware.OapiRequestValidator(spec)

	baseRouter := mux.NewRouter()
	apiRouter := baseRouter.PathPrefix("/api/v0/").Subrouter()
	apiRouter.Use(oapiValidator)

	api.HandlerFromMux(
		api.NewStrictHandler(server, []api.StrictMiddlewareFunc{}),
		apiRouter,
	)

	//nolint:exhaustruct
	httpServer := &http.Server{
		Addr:         "0.0.0.0:80",
		ReadTimeout:  serverCfg.ReadTimout,
		WriteTimeout: serverCfg.WriteTimout,
		Handler:      baseRouter,
	}

	logger.Info("started server")

	log.Fatal(httpServer.ListenAndServe())
}
