package main

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"runtime"

	_ "github.com/mattn/go-sqlite3"
)

const sqlitePragmaStuff = `
	pragma journal_mode = WAL;
	pragma synchronous = normal;
	pragma temp_store = memory;
	pragma mmap_size = 30000000000;
	pragma busy_timeout = 5000;
	pragma foreign_keys = ON;
`

//go:embed sql/create_table.sql
var createTableSQL string

func (s *Server) InitDatabase(ctx context.Context) error {
	writeDB, err := sql.Open("sqlite3", "file:zhteuern.sqlite")
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}

	if err := writeDB.Ping(); err != nil {
		return fmt.Errorf("db.Ping: %w", err)
	}

	// https://boyter.org/posts/go-sqlite-database-is-locked/
	writeDB.SetMaxOpenConns(1)

	_, err = writeDB.Exec(sqlitePragmaStuff)
	if err != nil {
		return fmt.Errorf("db pragma: %w", err)
	}

	_, err = writeDB.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("db table: %w", err)
	}

	readDB, err := sql.Open("sqlite3", "file:zhteuern.sqlite")
	if err != nil {
		return fmt.Errorf("sql.Open: %w", err)
	}

	if err := readDB.Ping(); err != nil {
		return fmt.Errorf("db.Ping: %w", err)
	}
	readDB.SetMaxOpenConns(runtime.NumCPU())

	s.readDB = &Queries{db: readDB}
	s.writeDB = &Queries{db: writeDB}
	s.closeDB = func() error {
		return errors.Join(readDB.Close(), writeDB.Close())
	}
	return nil
}
