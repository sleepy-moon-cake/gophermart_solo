package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseErrors(t *testing.T) {
	ctx := context.Background()
	_, err := DataBase(ctx, DbConfig{DSN: "invalid"})
	assert.Error(t, err)

	err = RunMigration(DbConfig{DSN: "invalid"})
	assert.Error(t, err)
}
