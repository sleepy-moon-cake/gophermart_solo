package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	secret := "test-secret-key"
	userID := 123

	// Helper to create a valid token
	validToken, _ := utils.BuildJwtToken(userID, secret)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		checkContext   bool
	}{
		{
			name:           "Valid Token",
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			checkContext:   true,
		},
		{
			name:           "Missing Header",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "Invalid Token Format",
			authHeader:     "Bearer invalid-token",
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
		{
			name:           "Wrong Secret Key",
			authHeader:     "Bearer " + func() string { t, _ := utils.BuildJwtToken(userID, "wrong"); return t }(),
			expectedStatus: http.StatusUnauthorized,
			checkContext:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Final handler to check if context was set
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					val := r.Context().Value(shared.UserID)
					assert.Equal(t, userID, val)
				}
				w.WriteHeader(http.StatusOK)
			})

			// Create middleware
			mw := AuthMiddleware(SetSecretKey(secret))
			handlerToTest := mw(nextHandler)

			// Execute request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()

			handlerToTest.ServeHTTP(rec, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestAuthMiddleware_Panic(t *testing.T) {
	// Ensure it panics if no secret key is provided
	assert.Panics(t, func() {
		AuthMiddleware()
	})
}
