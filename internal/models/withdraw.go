package models

import "time"

type WithdrawRequest struct {
	OrderNumber string  `json:"order_number"`
	Sum         float64 `json:"sum"`
}

type Withdraw struct {
	ID          int       `json:"-"`
	OwnerID     int       `json:"-"`
	OrderNumber string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type WithdrawWithCent struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}
