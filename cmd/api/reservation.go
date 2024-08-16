package main

import (
	"context"
	"fmt"
	"greenlight.zuyanh.net/internal/entity"
	"greenlight.zuyanh.net/internal/repository"
	"greenlight.zuyanh.net/internal/validator"
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

	user, err := app.models.Users.GetById(input.UserId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	total, err := app.models.Seat.UpdateSeatStatus(false, input.ShowId, input.SeatIds)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

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
		AppUser:       user.Email,
		ItemPrice:     strconv.FormatInt(total, 10),
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

func (app *application) listReservationHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserId int64
		ShowId int64
		Date   time.Time
		repository.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.UserId = int64(app.readInt(qs, "user_id", 0, v))
	input.ShowId = int64(app.readInt(qs, "show_id", 0, v))

	layout := "2006-01-02"
	dateStr := app.readString(qs, "created_at", "")
	input.Date, _ = time.Parse(layout, dateStr)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "amount", "user_id", "show_id", "created_at", "status", "-id", "-amount", "-user_id", "-show_id", "-created_at", "-status"}

	if repository.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	reservations, metadata, err := app.models.Reservation.GetAll(input.UserId, input.ShowId, input.Date, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"reservations": reservations, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func generateTransId(id int64) string {
	now := time.Now()
	return fmt.Sprintf("%02d%02d%02d_%v", now.Year()%100, int(now.Month()), now.Day(), id)
}
