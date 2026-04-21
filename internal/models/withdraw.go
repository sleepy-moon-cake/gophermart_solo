package models

import "time"

type WithdrawRequest struct {
	OrderNumber string  `json:"order_number"`
	Sum         float32 `json:"sum"`
}

type Withdraw struct {
	ID          int       `json:"-"`
	OwnerID     int       `json:"-"`
	OrderNumber string    `json:"order_number"`
	Sum         int       `json:"sum"`
	UploadedAt  time.Time `json:"-"`
}
