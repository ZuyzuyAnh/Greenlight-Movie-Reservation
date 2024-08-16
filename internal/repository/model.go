package repository

import (
	"database/sql"
	"errors"
	"time"

	"greenlight.zuyanh.net/internal/entity"
)

var (
	ErrRecordNotFound      = errors.New("record not found")
	ErrEditConflict        = errors.New("edit conflict")
	ErrDuplicateConstraint = errors.New("duplicate constraint")
	ErrViolatesForeignKey  = errors.New("violates foreign key constraint")
	ErrInvalidType         = errors.New("invalid type")
)

type Models struct {
	DB     *sql.DB
	Movies interface {
		Insert(movie *entity.Movie) error
		Get(id int64) (*entity.Movie, error)
		Update(movie *entity.Movie) error
		Delete(id int64) error
		GetAll(title string, genres []string, filters Filters) ([]*entity.Movie, Metadata, error)
	}
	Users interface {
		Insert(user *entity.User) error
		GetByEmail(email string) (*entity.User, error)
		Update(user *entity.User) error
		GetForToken(tokenScope, tokenPlaintext string) (*entity.User, error)
	}
	Token interface {
		New(userID int64, ttl time.Duration, scope string) (*entity.Token, error)
		Insert(token *entity.Token) error
		DeleteAllForUser(scope string, userID int64) error
	}
	Permissions interface {
		GetAllForUser(userID int64) (Permissions, error)
		AddForUser(userID int64, codes ...string) error
	}
	Theatres interface {
		Insert(theatres *entity.Theatres) error
		GetAll(city string, filters Filters) ([]*entity.Theatres, Metadata, error)
	}
	Screen interface {
		Insert(screen *entity.Screen) error
	}
	Seat interface {
		Insert(seat *entity.Seat) error
		InsertSeatStatus(showId int64, seatIds []int64) error
		GetAllByScreenId(screenId int64) ([]*entity.Seat, error)
		GetAllByShowId(showId int64, status string, filters Filters) ([]*entity.Seat, Metadata, error)
		UpdateSeatStatus(status bool, showId int64, seatId []int64) (int64, error)
	}
	Reservation interface {
		Insert(tx *sql.Tx, reservation *entity.Reservation, seatId []int64) error
		UpdateStatus(reservationId int64, status string) error
		GetById(id int64) (*entity.Reservation, error)
		Delete(id int64) error
	}
	Show interface {
		Insert(show *entity.Show) error
		GetAll(date string, title string, filters Filters) ([]*entity.Show, Metadata, error)
	}
}

func NewModel(db *sql.DB) Models {
	return Models{
		DB:          db,
		Movies:      MovieModel{DB: db},
		Users:       UserModel{DB: db},
		Token:       TokenModel{DB: db},
		Permissions: PermissionModel{DB: db},
		Theatres:    TheatresModel{DB: db},
		Screen:      ScreenModel{DB: db},
		Seat:        SeatModel{DB: db},
		Reservation: ReservationModel{DB: db},
		Show:        ShowModel{DB: db},
	}
}

// func NewMockModels() Models {
// 	return Models{
// 		Movies: MockMovieModel{},
// 	}
// }
