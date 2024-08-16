package main

import (
	"greenlight.zuyanh.net/internal/entity"
	"greenlight.zuyanh.net/internal/repository"
	"net/http"

	"greenlight.zuyanh.net/internal/validator"
)

func (app *application) createSeatHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Row       string `json:"row"`
		Number    int32  `json:"number"`
		Price     int32  `json:"price"`
		Screen_id int64  `json:"screen_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	seat := &entity.Seat{
		Row:       input.Row,
		Number:    input.Number,
		Price:     input.Price,
		Screen_id: input.Screen_id,
	}

	v := validator.New()

	if repository.ValidateSeat(v, seat); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Seat.Insert(seat)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"seat": seat}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) listAvailableSeatsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Show_id int64
		repository.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Show_id = int64(app.readInt(qs, "show_id", 0, v))

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "-id", "row", "-row", "number", "-number", "price", "-price"}

	if repository.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	seats, metadata, err := app.models.Seat.GetAllByShowId(input.Show_id, "", input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"seats": seats, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
