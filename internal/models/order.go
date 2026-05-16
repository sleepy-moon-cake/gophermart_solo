package models

import "time"

type Order struct {
	ID         int       `json:"-"`
	OwnerID    int       `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    int       `json:"accrual"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type OrderResponse struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"` // Будет null или отсутствовать, если nil
	UploadedAt time.Time `json:"uploaded_at"`
}

const (
	OrderStatusNew        = "NEW"        // Принят в обработку (202)
	OrderStatusProcessing = "PROCESSING" // В обработке
	OrderStatusInvalid    = "INVALID"    // Не прошел проверку
	OrderStatusProcessed  = "PROCESSED"  // Завершен
)
