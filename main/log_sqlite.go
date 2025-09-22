package main

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func NewLogger() (logger *slog.Logger, close func() error, err error) {
	writeDB, err := sql.Open("sqlite3", "file:log.sqlite")
	if err != nil {
		return nil, nil, fmt.Errorf("sql.Open: %w", err)
	}

	if err := writeDB.Ping(); err != nil {
		return nil, nil, fmt.Errorf("db.Ping: %w", err)
	}

	writeDB.SetMaxOpenConns(1)

	_, err = writeDB.Exec(sqlitePragmaStuff)
	if err != nil {
		return nil, nil, fmt.Errorf("db pragma: %w", err)
	}

	_, err = writeDB.Exec(createLogTableSQL)
	if err != nil {
		return nil, nil, fmt.Errorf("db table: %w", err)
	}

	sqliteLogger := &SQLiteLogger{
		Queries: &Queries{db: writeDB},
	}
	jsonLogger := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.LevelDebug,
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				a.Value = slog.TimeValue(time.Now().Truncate(time.Second).UTC())
			}
			return a
		},
	})

	var logHandler slog.Handler
	logHandler = &MultiLogger{
		handlers: []slog.Handler{sqliteLogger, jsonLogger},
	}
	logHandler = logHandler.WithAttrs([]slog.Attr{
		{Key: "app", Value: slog.StringValue(application)},
		{Key: "env", Value: slog.StringValue(environment)},
		{Key: "commit", Value: slog.StringValue(gitCommit)},
	})
	return slog.New(logHandler), writeDB.Close, nil
}

type SQLiteLogger struct {
	*Queries
}

//go:embed log_sqlite.sql
var createLogTableSQL string

func (h *SQLiteLogger) Enabled(context.Context, slog.Level) bool {
	return true
}

func (h *SQLiteLogger) Handle(ctx context.Context, r slog.Record) error {
	attrs := map[string]any{}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value.Any()
		return true
	})

	var attributeString string
	if len(attrs) > 0 {
		attrsByte, err := json.Marshal(attrs)
		if err != nil {
			return fmt.Errorf("attr json marshal: %w", err)
		}
		attributeString = string(attrsByte)
	}

	err := h.Queries.WriteLog(ctx, WriteLogParams{
		Msg:         r.Message,
		Attributes:  attributeString,
		LogLevel:    int(r.Level),
		TimeString:  r.Time.UTC().Format(time.RFC3339),
		TimeUnix:    int(r.Time.Unix()),
		App:         application,
		Env:         environment,
		CommitGit:   gitCommit,
		SrcFile:     r.Source().File,
		SrcLine:     r.Source().Line,
		SrcFunction: r.Source().Function,
	})
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}

	return nil
}

func (h *SQLiteLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *SQLiteLogger) WithGroup(name string) slog.Handler {
	return h
}

type MultiLogger struct {
	handlers []slog.Handler
}

func (h *MultiLogger) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *MultiLogger) Handle(ctx context.Context, r slog.Record) error {
	rCopy := r.Clone()
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, rCopy); err != nil {
			return err
		}
	}
	return nil
}

func (h *MultiLogger) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithAttrs(attrs)
	}
	return &MultiLogger{handlers: newHandlers}
}

func (h *MultiLogger) WithGroup(name string) slog.Handler {
	newHandlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		newHandlers[i] = handler.WithGroup(name)
	}
	return &MultiLogger{handlers: newHandlers}
}
