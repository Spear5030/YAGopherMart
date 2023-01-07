package domain

type User struct {
	ID      int64   `json:"id" db:"id"`
	Login   string  `json:"login" db:"login"`
	Hash    string  `json:"hash" db:"hash"`
	Balance float64 `json:"balance" db:"balance"`
}
