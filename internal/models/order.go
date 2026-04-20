package models

import "time"

type Order struct {
	ID         int       `json:"-"`
	OwnerID    int       `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
}

const (
	OrderStatusNew        = "NEW"        // Принят в обработку (202)
	OrderStatusProcessing = "PROCESSING" // В обработке
	OrderStatusInvalid    = "INVALID"    // Не прошел проверку
	OrderStatusProcessed  = "PROCESSED"  // Завершен
)
