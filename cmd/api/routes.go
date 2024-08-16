package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundErrorResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	//guest base
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	router.HandlerFunc(http.MethodGet, "/v1/theatres", app.listTheatreHandler)

	router.HandlerFunc(http.MethodGet, "/v1/shows", app.listShowHandler)

	router.HandlerFunc(http.MethodGet, "/v1/seats", app.listAvailableSeatsHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/callback", app.zaloPayCallBackHandler)

	//user base
	router.HandlerFunc(http.MethodPost, "/v1/payment", app.requirePermission("user", app.createReservationHandler))
	router.HandlerFunc(http.MethodGet, "/v1/reservations", app.requirePermission("user", app.listReservationHandler))

	//admin base
	router.HandlerFunc(http.MethodPost, "/v1/theatres", app.requirePermission("admin", app.createTheatreHandler))

	router.HandlerFunc(http.MethodPost, "/v1/screens", app.requirePermission("admin", app.createScreenHandler))

	router.HandlerFunc(http.MethodPost, "/v1/shows", app.requirePermission("admin", app.createShowHandler))
	router.HandlerFunc(http.MethodPost, "/v1/seats", app.requirePermission("admin", app.createSeatHandler))

	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("admin", app.createMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("admin", app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("admin", app.deleteMovieHandler))

	return app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))

}
