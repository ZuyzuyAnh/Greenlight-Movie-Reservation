package entity

type Seat struct {
	ID        int64  `json:"id"`
	Row       string `json:"row"`
	Number    int32  `json:"number"`
	Price     int32  `json:"price"`
	Screen_id int64  `json:"screen_id"`
}
