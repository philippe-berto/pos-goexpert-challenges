package repository

import (
	"context"
	"database/sql"
	"errors"
	q "server/quotes"
	"time"
)

const TimeoutError = "DB_OPERATION_TIMEOUT"

type (
	CreaterDollarQuote interface {
		CreateDollarQuote(c context.Context, quote q.DollarQuote, t time.Duration) error
	}

	Repository struct {
		ctx context.Context
		db  *sql.DB
	}
)

func New(ctx context.Context, db *sql.DB) *Repository {
	return &Repository{
		ctx: ctx,
		db:  db,
	}
}

func (r *Repository) CreateDollarQuote(c context.Context, quote q.DollarQuote, t time.Duration) error {
	ctx, cancel := context.WithTimeout(c, t)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `
    INSERT INTO dollar_quote (
      Code,
      Codein,
      Name,
      High,
      Low,
      VarBid,
      PctChange,
      Bid,
      Ask,
      Timestamp,
      CreateDate
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
  `, quote.Code, quote.Codein, quote.Name, quote.High, quote.Low, quote.VarBid, quote.PctChange, quote.Bid, quote.Ask, quote.Timestamp, quote.CreateDate)

	switch {
	case err == nil:
		return nil
	case ctx.Err() != nil:
		return errors.New(TimeoutError)
	default:
		return err
	}
}
