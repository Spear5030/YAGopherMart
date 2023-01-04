package domain

import "time"

type Order struct {
	Number     string    `json:"number" db:"number"`
	userID     int       `json:"userID" db:"user_id"`
	Status     string    `json:"status" db:"status"`
	Accrual    int64     `json:"accrual,omitempty" db:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at" db:"uploaded_at"`
	updatedAt  time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
