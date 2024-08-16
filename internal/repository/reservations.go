package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"greenlight.zuyanh.net/internal/entity"
	"time"
)

type ReservationModel struct {
	DB *sql.DB
}

func (m ReservationModel) Insert(tx *sql.Tx, reservation *entity.Reservation, seatId []int64) error {
	insertReservationQuery := `
		INSERT INTO reservations (user_id, amount, show_id) 
		VALUES ($1, $2, $3)
		RETURNING id, created_at, status
	`

	args := []interface{}{reservation.UserId, reservation.Amount, reservation.ShowId}

	err := tx.QueryRow(insertReservationQuery, args...).Scan(&reservation.ID, &reservation.CreatedAt, &reservation.Status)
	if err != nil {
		return err
	}

	insertReservationSeatsQuery := `
		INSERT INTO reservation_seat(reservation_id, seat_id)
		VALUES ($1, $2)
		RETURNING id
	`

	errors := make(chan error, len(seatId))
	defer close(errors)

	for _, seatId := range seatId {
		go func(seatId int64) {
			args := []interface{}{reservation.ID, seatId}
			_, err := tx.Exec(insertReservationSeatsQuery, args...)
			errors <- err
		}(seatId)
	}

	for i := 0; i < len(seatId); i++ {
		if err := <-errors; err != nil {
			return err
		}
	}
	return nil
}

func (m ReservationModel) UpdateStatus(reservationId int64, status string) error {
	query := `
		UPDATE reservations 
		SET status = $1
		WHERE id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{status, reservationId}

	_, err := m.DB.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m ReservationModel) GetById(id int64) (*entity.Reservation, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT id, user_id, amount, show_id, status
	FROM reservations
	WHERE id = $1
`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var reservation entity.Reservation

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&reservation.ID,
		&reservation.UserId,
		&reservation.Amount,
		&reservation.ShowId,
		&reservation.Status)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &reservation, nil
}

func (m ReservationModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
		DELETE FROM reservations
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m ReservationModel) GetAll(userId, showId int64, date time.Time, filters Filters) ([]*entity.Reservation, Metadata, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, user_id, amount, show_id, status
        FROM reservations
        WHERE (created_at::DATE = $1 OR $1 IS NULL) 
        AND (show_id = $2 OR $2 = 0)
        AND id = $3
        ORDER BY %s %s, id ASC
        LIMIT $4 OFFSET $5
	`, filters.sortColumn(), filters.sortDirection(),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{date, showId, userId, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0

	reservations := []*entity.Reservation{}
	for rows.Next() {
		var reservation entity.Reservation

		err := rows.Scan(
			&totalRecords,
			&reservation.ID,
			&reservation.CreatedAt,
			&reservation.UserId,
			&reservation.Amount,
			&reservation.ShowId,
			&reservation.Status,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		reservations = append(reservations, &reservation)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return reservations, metadata, nil
}
