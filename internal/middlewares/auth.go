package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/sleepy-moon-cake/gophermart_solo/internal/shared"
	"github.com/sleepy-moon-cake/gophermart_solo/internal/utils"
)

type jwtParams struct {
	expired time.Duration
	secretKey string
}

type Options func(*jwtParams)



func SetSecretKey(secretKey string) Options{
	return func (p *jwtParams)  {
		p.secretKey = secretKey
	}
}

func AuthMiddleware (opts ...Options)  func(http.Handler) http.Handler {
	p:= &jwtParams{
		expired: time.Hour * 24,
	}

	for _,opt:= range opts {
		opt(p)
	}

	auth:= auth{
		jwtParams: *p,
	}

	if p.secretKey == ""{
		panic("auth middleware: secret key is required")
	}

	return auth.authMiddleware
} 

type auth struct {
	jwtParams
}

func (auth *auth) authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headerValue:=r.Header.Get("Authorization")

		if headerValue == ""{
			http.Error(w,http.StatusText(http.StatusUnauthorized),http.StatusUnauthorized)
			return  
		}

		userId,err:=utils.ParseJwtToken(headerValue,auth.secretKey)

		if err !=nil {
			http.Error(w,http.StatusText(http.StatusUnauthorized),http.StatusUnauthorized)
			return  
		}

		ctx:=context.WithValue(r.Context(), shared.UserID,userId)

		r = r.WithContext(ctx)	

		h.ServeHTTP(w,r)
	})
}

