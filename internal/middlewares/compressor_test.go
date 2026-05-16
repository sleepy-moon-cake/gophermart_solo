package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressor(t *testing.T) {
	// Базовый хендлер, который возвращает JSON
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Если это POST, прочитаем тело и отправим обратно (echo)
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			w.Write(body)
			return
		}
		w.Write([]byte(`{"status":"ok"}`))
	})

	mw := Compressor(handler)

	t.Run("Should compress response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "gzip", rec.Header().Get("Content-Encoding"))

		// Проверяем, что данные реально сжаты и их можно прочитать
		zr, err := gzip.NewReader(rec.Body)
		require.NoError(t, err)
		defer zr.Close()

		decodedBody, _ := io.ReadAll(zr)
		assert.JSONEq(t, `{"status":"ok"}`, string(decodedBody))
	})

	t.Run("Should decompress request", func(t *testing.T) {
		// Сжимаем данные для отправки
		data := `{"test":"data"}`
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		zw.Write([]byte(data))
		zw.Close()

		req := httptest.NewRequest(http.MethodPost, "/", &buf)
		req.Header.Set("Content-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		// Если декомпрессия сработала, хендлер вернул нам чистый JSON
		assert.JSONEq(t, data, rec.Body.String())
	})

	t.Run("Should NOT compress if no header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.Empty(t, rec.Header().Get("Content-Encoding"))
		assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
	})
}
