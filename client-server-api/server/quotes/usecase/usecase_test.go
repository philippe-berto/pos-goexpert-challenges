package usecase

import (
	"context"
	"testing"

	"github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/quotes/repository"

	mydb "github.com/philippe-berto/pos-goexpert-challenges/client-server-api/server/database"

	"github.com/stretchr/testify/assert"
)

func TestUsecase(t *testing.T) {
	t.Run("should get dollar quote", func(t *testing.T) {
		ctx := context.Background()
		db, err := mydb.New(ctx, mydb.Config{File: "sqlite.s3db", RunMigration: true})
		assert.NoError(t, err)

		repo := repository.New(ctx, db.GetConnection())
		usecase := New(ctx, repo, Config{ApiCallTimeoutMs: 2000, DbOperationTimeoutMs: 10})

		dollarQuote, err := usecase.GetDollarQuote()
		if err != nil {
			t.Errorf("getDollarQuote() error = %v, want nil", err)
		}

		assert.NotEmpty(t, dollarQuote)
		assert.NoError(t, err)
	})

	t.Run("should get EXTERNAL_API_CALL_TIMEOUT", func(t *testing.T) {
		ctx := context.Background()
		db, err := mydb.New(ctx, mydb.Config{File: "sqlite.s3db", RunMigration: true})
		assert.NoError(t, err)

		repo := repository.New(ctx, db.GetConnection())
		usecase := New(ctx, repo, Config{ApiCallTimeoutMs: 2, DbOperationTimeoutMs: 10})

		_, err = usecase.GetDollarQuote()
		assert.Error(t, err)
		assert.Equal(t, TimeoutError, err.Error())

	})

	t.Run("should get DB_OPERATION_TIMEOUT", func(t *testing.T) {
		ctx := context.Background()
		db, err := mydb.New(ctx, mydb.Config{File: "sqlite.s3db", RunMigration: true})
		assert.NoError(t, err)

		repo := repository.New(ctx, db.GetConnection())
		usecase := New(ctx, repo, Config{ApiCallTimeoutMs: 200, DbOperationTimeoutMs: 0.001})

		_, err = usecase.GetDollarQuote()
		assert.Error(t, err)
		assert.Equal(t, "DB_OPERATION_TIMEOUT", err.Error())
	})
}
