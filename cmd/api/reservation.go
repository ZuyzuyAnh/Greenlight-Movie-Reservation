package main

import (
	"context"
	"fmt"
	"greenlight.zuyanh.net/internal/entity"
	"net/http"
	"strconv"
	"time"
)

func (app *application) createReservationHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Minute)
	defer cancel()

	tx, err := app.models.DB.BeginTx(ctx, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	defer tx.Commit()

	var input struct {
		UserId  int64   `json:"user_id"`
		ShowId  int64   `json:"show_id"`
		SeatIds []int64 `json:"seat_ids"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		tx.Rollback()
		return
	}

	total, err := app.models.Seat.UpdateSeatStatus(false, input.ShowId, input.SeatIds)

	reservation := &entity.Reservation{
		UserId: input.UserId,
		ShowId: input.ShowId,
		Amount: total,
	}

	err = app.models.Reservation.Insert(tx, reservation, input.SeatIds)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		tx.Rollback()
		return
	}

	param := Params{
		AppUser:       strconv.FormatInt(input.UserId, 10),
		ItemPrice:     "10000",
		ReservationId: generateTransId(reservation.ID),
	}

	response, err := CreaterOrder(param)
	fmt.Println(response)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		tx.Rollback()
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"payment": response}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		tx.Rollback()
		return
	}

	go func(seatIds []int64) {
		time.Sleep(10 * time.Minute)

		res, err := app.models.Reservation.GetById(reservation.ID)
		if err != nil {

		}
		if res.Status == "pending" {
			go func() {
				_, err = app.models.Seat.UpdateSeatStatus(true, reservation.ShowId, seatIds)
				if err != nil {
					app.logger.PrintError(err, nil)
				}
			}()
			err = app.models.Reservation.Delete(res.ID)
			if err != nil {
				app.logger.PrintError(err, nil)
			}
		}
	}(input.SeatIds)
}

func generateTransId(id int64) string {
	now := time.Now()
	return fmt.Sprintf("%02d%02d%02d_%v", now.Year()%100, int(now.Month()), now.Day(), id)
}
