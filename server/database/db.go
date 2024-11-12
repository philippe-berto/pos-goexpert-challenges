package db

import (
	"context"
	"database/sql"

	_ "github.com/caarlos0/env"
	_ "github.com/mattn/go-sqlite3"

	"log"
)

type (
	Config struct {
		File         string `env:"DB_FILE" envDefault:"sqlite.s3db"`
		RunMigration bool   `env:"DB_MIGRATION" envDefault:"true"`
	}

	Client struct {
		ctx context.Context
		db  *sql.DB
		cfg Config
	}
)

func New(ctx context.Context, cfg Config) (*Client, error) {

	db, err := sql.Open("sqlite3", cfg.File)
	if err != nil {
		log.Fatal(err)
	}
	err = Migrate(cfg, db)
	if err != nil {
		log.Fatal(err)
	}
	return &Client{
		ctx: ctx,
		db:  db,
		cfg: cfg,
	}, nil
}

func (c *Client) GetConnection() *sql.DB {
	return c.db
}

func (c *Client) Close() error {
	return c.db.Close()
}

func Migrate(cfg Config, db *sql.DB) error {
	if !cfg.RunMigration {
		log.Println("skipping migration")
		return nil
	}
	log.Println("running migration")
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS dollar_quote (
      Code TEXT,
      Codein TEXT,
      Name TEXT,
      High TEXT,
      Low TEXT,
      VarBid TEXT,
      PctChange TEXT,
      Bid TEXT,
      Ask TEXT,
      Timestamp TEXT,
      CreateDate TEXT
    );
  `)
	return err
}
