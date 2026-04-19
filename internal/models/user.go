package models

type LoginParams = RegisterRequest

type LoginRequest = RegisterRequest

type RegisterData = RegisterRequest

type RegisterParams struct {
	Login string
	HashPassword string
}
type RegisterRequest struct {
	Login string `json:"login"`
	Password string `json:"password"`
}

type User struct {
	ID int `json:"id"`
	Login string `json:"login"`
	Password string `json:"password"`
}