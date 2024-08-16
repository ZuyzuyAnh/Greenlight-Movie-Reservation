package entity

import "time"

type Reservation struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Amount    int64     `json:"amount"`
	Status    string    `json:"status"`
	UserId    int64     `json:"user_id"`
	ShowId    int64     `json:"show_id"`
}
