package repository

import (
	"context"
	"database/sql"
	"greenlight.zuyanh.net/internal/entity"
	"time"

	"github.com/lib/pq"
	"greenlight.zuyanh.net/internal/validator"
)

type ScreenModel struct {
	DB *sql.DB
}

func (m ScreenModel) Insert(screen *entity.Screen) error {
	query := `
		INSERT INTO screens(number, theatre_id)
		VALUES ($1, $2)
		RETURNING id
	`

	args := []interface{}{
		screen.Number,
		screen.Theatre_id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&screen.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23503" {
				return ErrViolatesForeignKey
			}
		} else {
			return err
		}
	}

	return nil
}

func ValidateScreens(v *validator.Validator, screen *entity.Screen) {
	v.Check(screen.Number < 100, "number", "must be less than 100")

	v.Check(screen.Number > 0, "number", "must be a positive integer")
	v.Check(screen.Theatre_id > 0, "theatre_id", "must be a positive integer")
}
