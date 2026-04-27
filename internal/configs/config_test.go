package configs

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfig(t *testing.T) {
	os.Setenv("RUN_ADDRESS", "localhost:9999")
	cfg := GetConfig()
	assert.Equal(t, "localhost:9999", cfg.ServerAddress)
}
