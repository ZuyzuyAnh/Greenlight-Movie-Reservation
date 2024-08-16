package main

import (
	"greenlight.zuyanh.net/internal/entity"
	"greenlight.zuyanh.net/internal/repository"
	"net/http"

	"greenlight.zuyanh.net/internal/validator"
)

func (app *application) createTheatreHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string `json:"name"`
		City string `json:"city"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	theatre := &entity.Theatres{
		Name: input.Name,
		City: input.City,
	}

	v := validator.New()

	if repository.ValidateTheatres(v, theatre); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Theatres.Insert(theatre)
	if err != nil {
		switch err {
		case repository.ErrDuplicateConstraint:
			app.duplicateConstraintResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"theatre": theatre}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listTheatreHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		City string
		repository.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.City = app.readString(qs, "city", "")

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "-id", "name", "-name"}

	if repository.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	theatres, metadata, err := app.models.Theatres.GetAll(input.City, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"theatres": theatres, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
