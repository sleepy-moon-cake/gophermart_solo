package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggerMiddleware(t *testing.T) {
	t.Run("Capture status code and continue execution", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("ok"))
		})

		handlerToTest := LoggerMiddleware(nextHandler)

		req := httptest.NewRequest(http.MethodPost, "/test-uri", nil)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})

	t.Run("Default status code should be 200", func(t *testing.T) {
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("no status set"))
		})

		handlerToTest := LoggerMiddleware(nextHandler)

		req := httptest.NewRequest(http.MethodGet, "/default", nil)
		rec := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})
}

func TestResponseWriterWrapper(t *testing.T) {
	rec := httptest.NewRecorder()
	ww := &ResponseWriterWrapper{ResponseWriter: rec, statusCode: http.StatusOK}

	ww.WriteHeader(http.StatusNotFound)

	assert.Equal(t, http.StatusNotFound, ww.statusCode)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
