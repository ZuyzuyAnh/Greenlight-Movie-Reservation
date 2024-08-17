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

type TheatresModel struct {
	DB *sql.DB
}

func (m TheatresModel) Insert(theatres *entity.Theatres) error {
	query := `
		INSERT INTO theatres(name, city)
		VALUES ($1, $2)
		RETURNING id
	`
	args := []interface{}{
		theatres.Name,
		theatres.City,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&theatres.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" {
				return ErrDuplicateConstraint
			}
		} else {
			return err
		}
	}

	return nil
}

func (m TheatresModel) GetAll(city string, filters Filters) ([]*entity.Theatres, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, name, city
		FROM theatres
		WHERE (to_tsvector('simple', city) @@ plainto_tsquery('simple', $1) OR $1 = '') 
		ORDER BY %s %s, id ASC
		LIMIT $2 OFFSET $3
	`, filters.sortColumn(), filters.sortDirection(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{city, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0

	theatres := []*entity.Theatres{}
	for rows.Next() {
		var theatre entity.Theatres

		err := rows.Scan(
			&totalRecords,
			&theatre.ID,
			&theatre.Name,
			&theatre.City,
		)

		if err != nil {
			return nil, Metadata{}, err
		}

		theatres = append(theatres, &theatre)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return theatres, metadata, nil
}

func (m TheatresModel) Update(theatres *entity.Theatres) error {
	query := `
		UPDATE theatres
		SET name = $1, city = $2
		WHERE id = $3
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{
		theatres.Name,
		theatres.City,
		theatres.ID,
	}

	_, err := m.DB.ExecContext(ctx, query, args)

	if err != nil {
		return err
	}

	return nil
}

func (m TheatresModel) Delete(theatreId int64) error {
	query := `
		DELETE FROM theatres
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, theatreId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func ValidateTheatres(v *validator.Validator, theatres *entity.Theatres) {
	v.Check(theatres.Name != "", "name", "must be provided")
	v.Check(len(theatres.Name) < 100, "name", "must not be more than 100 characters")

	v.Check(theatres.City != "", "city", "must be provided")
	v.Check(len(theatres.City) < 100, "city", "must not be more than 100 characters")
}
