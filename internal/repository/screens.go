package repository

import (
	"context"
	"database/sql"
	"fmt"
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

func (m ScreenModel) GetAll(theatreId int64, filters Filters) ([]*entity.Screen, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, number, theatre_id
		FROM screens
		WHERE theatre_id = $1 OR $1 = 0
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3
	`, filters.sortColumn(), filters.sortDirection(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{theatreId, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0

	screens := []*entity.Screen{}
	for rows.Next() {
		var screen entity.Screen

		err := rows.Scan(
			&totalRecords,
			&screen.ID,
			&screen.Number,
			&screen.Theatre_id,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		screens = append(screens, &screen)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return screens, metadata, nil
}

func (m ScreenModel) Update(screen *entity.Screen) error {
	query := `
		UPDATE screens
		SET number = $1, theatre_id = $2
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		screen.Number,
		screen.Theatre_id,
		screen.ID,
	}

	_, err := m.DB.ExecContext(ctx, query, args)

	if err != nil {
		return err
	}

	return nil
}

func (m ScreenModel) Delete(screenId int64) error {
	query := `
		DELETE FROM screens
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, screenId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateScreens(v *validator.Validator, screen *entity.Screen) {
	v.Check(screen.Number < 100, "number", "must be less than 100")

	v.Check(screen.Number > 0, "number", "must be a positive integer")
	v.Check(screen.Theatre_id > 0, "theatre_id", "must be a positive integer")
}
