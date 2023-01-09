package domain

import "time"

type Order struct {
	Number     string    `json:"number" db:"number"`
	userID     int       `db:"user_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    float64   `json:"accrual,omitempty" db:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
	updatedAt  time.Time `db:"updated_at"`
}
