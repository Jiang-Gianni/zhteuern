package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
)

// These variables will be set by -ldflags during the compilation and printed out in main
// go build -ldflags="-X main.environment=${ENV} -X main.gitCommit=${GIT_COMMIT}" -o bin/main ./main/*.go
var (
	environment = "dev"
	gitCommit   = "gitCommit"
	application = "zhteuern"
	host        = "zhteuern.giannijiang.vip"
	port        = "3000"
)

func main() {
	slog.Info("main", "env", environment, "commit", gitCommit)
	if err := run(); err != nil {
		slog.Error("run", "error", err)
		os.Exit(1)
	}
}

func run() (err error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slogger, closeLog, err := NewLogger()
	if err != nil {
		return fmt.Errorf("new logger: %w", err)
	}
	defer func() {
		slogger.Info("STOP")
		closeLogErr := closeLog()
		if closeLogErr != nil {
			err = errors.Join(err, closeLogErr)
		}
	}()

	s := &Server{
		Environment: environment,
		Port:        port,
		Logger:      slogger,
		Host:        host,
	}
	s.RegisterHttp()

	if err := s.InitDatabase(ctx); err != nil {
		return fmt.Errorf("init db: %w", err)
	}

	slogger.Info("START", "port", s.Port)
	return s.Run(ctx)
}
