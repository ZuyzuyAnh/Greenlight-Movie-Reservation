package main

import (
	"errors"
	"fmt"
	"greenlight.zuyanh.net/internal/entity"
	"greenlight.zuyanh.net/internal/repository"
	"net/http"
	"time"

	"greenlight.zuyanh.net/internal/validator"
)

func (app *application) createShowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ShowTime time.Time `json:"showtime"`
		MovieId  int64     `json:"movie_id"`
		ScreenId int64     `json:"screen_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	seats, err := app.models.Seat.GetAllByScreenId(input.ScreenId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	seatIds := make([]int64, len(seats))

	for i, seat := range seats {
		seatIds[i] = seat.ID
	}
	fmt.Printf("Seats retrieved: %v\n", seatIds)

	show := &entity.Show{
		Showtime: input.ShowTime,
		MovieId:  input.MovieId,
		ScreenId: input.ScreenId,
	}

	v := validator.New()

	if repository.ValidateShow(v, show); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Show.Insert(show)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrViolatesForeignKey):
			app.violateForeignKeyResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.models.Seat.InsertSeatStatus(show.ID, seatIds)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"show": show}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listShowHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Date  string
		Title string
		repository.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Date = app.readString(qs, "date", "")
	input.Title = app.readString(qs, "title", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "-id", "datetime", "-datetime"}

	repository.ValidateDateFormat(v, input.Date)
	repository.ValidateFilters(v, input.Filters)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	shows, metadata, err := app.models.Show.GetAll(input.Date, input.Title, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"shows": shows, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
