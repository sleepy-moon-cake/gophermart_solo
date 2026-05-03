package middlewares

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ResponseWriterWrapper struct{
	http.ResponseWriter
	statusCode int
}

func (ww * ResponseWriterWrapper) WriteHeader(statusCode int){
	ww.statusCode = statusCode
	ww.ResponseWriter.WriteHeader(statusCode)
}


func LoggerMiddleware( h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start:= time.Now()

		ww:=&ResponseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		defer func (){
			slog.Info("Request", 
				slog.String("Method", r.Method), 
				slog.String("RequestURI", r.RequestURI),
				slog.Duration("Duration", time.Since(start)),
				slog.Int("StatusCode", ww.statusCode),
			)
			
			fmt.Println("")
		}()
		

		h.ServeHTTP(ww,r)
	})
}