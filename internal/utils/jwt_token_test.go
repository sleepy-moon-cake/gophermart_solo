package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJwtTokens(t *testing.T) {
	secret := "my-secret-key"
	userID := 42

	t.Run("Build and Parse Valid Token", func(t *testing.T) {
		token, err := BuildJwtToken(userID, secret)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		parsedID, err := ParseJwtToken(token, secret)
		assert.NoError(t, err)
		assert.Equal(t, userID, parsedID)
	})

	t.Run("Fail with Wrong Secret", func(t *testing.T) {
		token, _ := BuildJwtToken(userID, secret)

		_, err := ParseJwtToken(token, "different-secret")
		assert.Error(t, err)
	})

	t.Run("Fail with Invalid Token String", func(t *testing.T) {
		_, err := ParseJwtToken("not-a-token", secret)
		assert.Error(t, err)
	})

	t.Run("Fail with Expired Token", func(t *testing.T) {
		// Manually create an expired token
		claims := Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 1 hour ago
			},
			UserId: userID,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredString, _ := token.SignedString([]byte(secret))

		_, err := ParseJwtToken(expiredString, secret)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "token is expired")
	})
}
