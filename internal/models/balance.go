package models

type Balance struct {
	ID        int `json:"-"`
	OwnerID   int `json:"-"`
	Current   int `json:"current"`
	Withdrawn int `json:"withdrawn"`
}
