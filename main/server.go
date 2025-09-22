package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Environment string
	Port        string
	Logger      *slog.Logger
	HttpServer  *http.Server
	Host        string

	/*
		http.Server.Shutdown will hang until ctx's timeout
		because of SSE's active connection.
		closeIdleConns() in stdlib's net/http.server.go:
		quiescent is always false for active SSE's connection.
		shutdownChan is closed on server shutdown so that all channel
		listeners will receive the (zero, false) value (ex in SSE handlers)
	*/
	shutdownChan chan struct{}
	readDB       *Queries
	writeDB      *Queries
	closeDB      func() error
}

func (s *Server) RegisterHttp() {
	mux := http.NewServeMux()
	patternToHttpHandler := map[string]httpHandlerFunc{
		"GET /brotli/":                       s.GetAssetsHandler(),
		"GET /not-found":                     s.GetNotFoundHandler(),
		"/":                                  s.IndexHandler(),
		"/tax-simulation/{tsID}":             s.TaxSimulationHandler(),
		"GET /tax-simulation/{tsID}/qr-code": s.TaxSimulationQRCodeHandler(),
	}
	if s.Environment == EnvironmentDEV {
		patternToHttpHandler["PATCH /hot"] = s.HotReloadHandler()
	}
	pageNotFoundBrotli, err := s.BrotliTempl(ViewPageNotFound())
	if err != nil {
		log.Panicf("brotli: %v", err)
	}
	internalErrorBrotli, err := s.BrotliTempl(ViewInternalError())
	if err != nil {
		log.Panicf("brotli: %v", err)
	}
	for pattern, handler := range patternToHttpHandler {
		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			next := s.AllowedRequest(handler)
			if err := next(w, r); err != nil {
				switch {
				case errors.Is(err, context.Canceled),
					errors.Is(err, syscall.EPIPE),
					errors.Is(err, context.DeadlineExceeded),
					err.Error() == "sql: database is closed":
					// skip, don't care
				case errors.Is(err, ErrPageNotFound),
					errors.Is(err, sql.ErrNoRows):
					pageNotFoundBrotli(w, r)
				case errors.Is(err, ErrForbidden):
					http.Error(w, "forbidden", http.StatusForbidden)
				case errors.Is(err, ErrMethodNotAllowed):
					http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				default:
					http.Error(w, "internal server error", http.StatusInternalServerError)
					internalErrorBrotli(w, r)
					s.Logger.Error("http", "err", err, "method", r.Method, "path", r.URL.Path)
				}
			}
		})
	}
	s.HttpServer = &http.Server{
		Handler:           mux,
		Addr:              fmt.Sprintf(":%s", s.Port),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}
	s.shutdownChan = make(chan struct{})
	s.HttpServer.RegisterOnShutdown(func() {
		close(s.shutdownChan)
	})
}

func (s *Server) Run(ctx context.Context) (err error) {
	defer func() {
		err = errors.Join(err, s.closeDB())
		if panicErr := recover(); panicErr != nil {
			err = fmt.Errorf("%w: %v", err, panicErr)
		}
	}()

	errChan := make(chan error)
	go func() {
		errChan <- s.HttpServer.ListenAndServe()
	}()

	go func() {
		errChan <- s.GracefulShutdown(ctx)
	}()

	for err := range errChan {
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}

func (s *Server) GracefulShutdown(ctx context.Context) error {
	sigChn := make(chan os.Signal, 1)
	signal.Notify(sigChn, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-sigChn
	timeout := time.Second * 5
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return s.HttpServer.Shutdown(ctx)
}
