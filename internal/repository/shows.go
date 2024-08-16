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

type ShowModel struct {
	DB *sql.DB
}

func (m ShowModel) Insert(show *entity.Show) error {
	insertShowQuery := `
		INSERT INTO shows(showtime, movie_id, screen_id)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	args := []interface{}{
		show.Showtime,
		show.MovieId,
		show.ScreenId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, insertShowQuery, args...).Scan(&show.ID)
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

func (m ShowModel) GetAll(date string, title string, filters Filters) ([]*entity.Show, Metadata, error) {
	if date != "" {
		_, err := time.Parse("2006-01-02", date)
		if err != nil {
			return nil, Metadata{}, ErrInvalidType
		}
	}

	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(), 
               s.id, 
               s.showtime, 
               s.movie_id,
               s.screen_id
		FROM show s
		INNER JOIN movies m ON s.movie_id = m.id
		WHERE (to_tsvector('simple', m.title) @@ plainto_tsquery('simple', $1) OR $1 = '')
		  AND (TO_CHAR(s.showtime::date, 'YYYY-MM-DD') = $2 OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4;
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, date, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0

	shows := []*entity.Show{}
	for rows.Next() {
		var show entity.Show

		err := rows.Scan(
			&totalRecords,
			&show.ID,
			&show.Showtime,
			&show.MovieId,
			&show.ScreenId,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		shows = append(shows, &show)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return shows, metadata, nil
}

func ValidateShow(v *validator.Validator, show *entity.Show) {
	v.Check(show.MovieId > 0, "movie_id", "must be a positive integer")
	v.Check(show.ScreenId > 0, "screen_id", "must be a positive integer")
}
