package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"greenlight.zuyanh.net/internal/entity"
	"sync"
	"time"
	"unicode"

	"github.com/lib/pq"
	"greenlight.zuyanh.net/internal/validator"
)

type SeatModel struct {
	DB *sql.DB
}

func (m SeatModel) Insert(seat *entity.Seat) error {
	query := `
		INSERT INTO seats(row, number, price, screen_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	args := []interface{}{
		seat.Row,
		seat.Number,
		seat.Price,
		seat.Screen_id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&seat.ID)
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

func (m SeatModel) InsertSeatStatus(showId int64, seatIds []int64) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(seatIds))

	errors := make(chan error, len(seatIds))

	query := `
		INSERT INTO seat_status(seat_id, show_id)
		VALUES ($1, $2)
	`

	for _, seatId := range seatIds {
		go func(seatId int64) {
			defer wg.Done()
			_, err := m.DB.ExecContext(ctx, query, seatId, showId)
			if err != nil {
				errors <- err
			}
		}(seatId)
	}

	go func() {
		wg.Wait()
		close(errors)
	}()
	for err := range errors {
		if err != nil {
			return err
		}
	}

	return nil
}

func (m SeatModel) GetAllByScreenId(screenId int64) ([]*entity.Seat, error) {
	query := `
		SELECT id, row, number, price, screen_id
		FROM seats
		WHERE screen_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, screenId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seats := []*entity.Seat{}
	for rows.Next() {
		var seat entity.Seat

		err := rows.Scan(
			&seat.ID,
			&seat.Row,
			&seat.Number,
			&seat.Price,
			&seat.Screen_id,
		)
		if err != nil {
			return nil, err
		}

		seats = append(seats, &seat)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return seats, nil
}

func (m SeatModel) GetAllByShowId(showId int64, status string, filters Filters) ([]*entity.Seat, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT COUNT(*) OVER(),
			s.id,
			s.row,
			s.number,
			s.price,
			s.screen_id
		FROM seats s 
		INNER JOIN shows sh ON s.screen_id = sh.screen_id
		INNER JOIN seat_status sst ON s.id = sst.seat_id
		WHERE sh.show_id = $1 AND (sst.status = $2 OR $2 = '')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4
	`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{showId, status, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0

	seats := []*entity.Seat{}
	for rows.Next() {
		var seat entity.Seat

		err := rows.Scan(
			&totalRecords,
			&seat.ID,
			&seat.Row,
			&seat.Number,
			&seat.Price,
			&seat.Screen_id,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		seats = append(seats, &seat)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return seats, metadata, nil

}

func (m SeatModel) UpdateSeatStatus(status bool, showId int64, seatId []int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if len(seatId) == 0 {
		return 0, errors.New("no seat IDs provided")
	}

	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	seatLockQuery := `
		SELECT *
		FROM seat_status
		WHERE seat_id = ANY($1)
		FOR UPDATE
	`

	priceCalQuery := `
		SELECT SUM(price)
		FROM seats
		WHERE id = ANY($1)
	`

	type result struct {
		price int64
		err   error
	}

	results := make(chan result, 2)
	defer close(results)

	go func() {
		_, err := tx.ExecContext(ctx, seatLockQuery, pq.Array(seatId))
		results <- result{err: err}
	}()

	var total int64

	go func() {
		err := tx.QueryRowContext(ctx, priceCalQuery, pq.Array(seatId)).Scan(&total)
		results <- result{price: total, err: err}
	}()

	for i := 0; i < 2; i++ {
		res := <-results
		if res.err != nil {
			tx.Rollback()
			return 0, res.err
		}
	}

	queryUpdateSeatStatus := `
		UPDATE seat_status
		SET available = $1 
		WHERE show_id = $2 AND seat_id = ANY($3)
	`

	_, err = tx.ExecContext(ctx, queryUpdateSeatStatus, status, showId, pq.Array(seatId))
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	err = tx.Commit()
	if err != nil {
		return 0, err
	}

	return total, nil
}

func ValidateSeat(v *validator.Validator, seat *entity.Seat) {
	v.Check(seat.Row != "", "row", "must be provided")
	v.Check(unicode.IsLetter(rune(seat.Row[0])), "row", "must be a alphabet")

	v.Check(seat.Number > 0, "number", "must be a positive integer")
	v.Check(seat.Price > 0, "price", "must be a positive integer")
	v.Check(seat.Screen_id > 0, "screen_id", "must be a positive integer")
}
