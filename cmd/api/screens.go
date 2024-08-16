package main

import (
	"greenlight.zuyanh.net/internal/entity"
	data "greenlight.zuyanh.net/internal/repository"
	"net/http"

	"greenlight.zuyanh.net/internal/validator"
)

func (app *application) createScreenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Number    int32 `json:"number"`
		TheatreId int64 `json:"theatre_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	screen := &entity.Screen{
		Number:     input.Number,
		Theatre_id: input.TheatreId,
	}

	v := validator.New()

	if data.ValidateScreens(v, screen); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Screen.Insert(screen)
	if err != nil {
		switch err {
		case data.ErrViolatesForeignKey:
			app.violateForeignKeyResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"screen": screen}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
