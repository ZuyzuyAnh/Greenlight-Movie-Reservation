package entity

import "time"

type Show struct {
	ID       int64     `json:"id"`
	Showtime time.Time `json:"showtime"`
	MovieId  int64     `json:"movie_id"`
	ScreenId int64     `json:"screen_id"`
}
