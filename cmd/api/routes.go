package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	// router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
	//router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))

	// router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
	// router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
	// router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/perusahaans", app.listPerusahaanHandler)
	router.HandlerFunc(http.MethodGet, "/v1/warehouse", app.listWarehouseHandler)
	router.HandlerFunc(http.MethodGet, "/v1/rak", app.listRakHandler)
	router.HandlerFunc(http.MethodGet, "/v1/brand", app.listBrandHandler)
	router.HandlerFunc(http.MethodGet, "/v1/brandasset", app.listBrandAssetHandler)
	router.HandlerFunc(http.MethodGet, "/v1/stok", app.listStokHandler)
	//listStokHandler

	router.HandlerFunc(http.MethodPost, "/v1/perusahaans", app.createPerusahaanHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodPost, "/v1/warehouse", app.createWarehouseHandler)
	router.HandlerFunc(http.MethodPost, "/v1/rak", app.createRakHandler)
	router.HandlerFunc(http.MethodPost, "/v1/brand", app.createBrandHandler)
	router.HandlerFunc(http.MethodPost, "/v1/brandasset", app.createBrandAssetHandler)
	router.HandlerFunc(http.MethodPost, "/v1/stok", app.createStokHandler)
	//createStokHandler

	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/rak/:id", app.showRakHandler)
	router.HandlerFunc(http.MethodGet, "/v1/stok/:id", app.showStokHandler)
	//router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/perusahaans/:id", app.updatePerusahaanHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/warehouse/:id", app.updateWarehouseHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/rak/:id", app.updateRakHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/brand/:id", app.updateBrandHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/brandasset/:id", app.updateBrandAssetHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/stok/:id", app.updateStokHandler)
	//updateStokHandler

	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/perusahaans/:id", app.deletePerusahaanHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/warehouse/:id", app.deleteWarehouseHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/rak/:id", app.deleteRakHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/brand/:id", app.deleteBrandHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/brandasset/:id", app.deleteBrandAssetHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/stok/:id", app.deleteStokHandler)

	//deleteStokHandler

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
