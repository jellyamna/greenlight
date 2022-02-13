package main

import (
	"errors"
	"net/http"
	"time"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

// var wg sync.WaitGroup

func (app *application) createRakHandler(w http.ResponseWriter, r *http.Request) {

	// type input struct {
	// 	Rak_code     string `json:"rak_code"`
	// 	Rak_ket      string `json:"rak_ket"`
	// 	Warehouse_id int32  `json:"warehouse_id"`
	// }

	RakMulti := []data.RakMultiInsert{}

	err := app.readJSONArray(w, r, &RakMulti)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Rak.Insert(&RakMulti)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) showRakHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Rak.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rak": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateRakHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Rak.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Rak_id         *int64     `json:"rak_id"`
		Created_at     *time.Time `json:"created_at"`
		Rak_code       *string    `json:"rak_code"`
		Rak_ket        *string    `json:"rak_ket"`
		Version        *int32     `json:"version"`
		Modified_at    *time.Time `json:"modified_at"`
		User_modified  *string    `json:"user_modified"`
		Warehouse_id   *int32     `json:"warehouse_id"`
		Name_warehouse *string    `json:"Name_warehouse"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	// usaha.Rak_code,
	// 	usaha.Rak_ket,
	// 	usaha.Warehouse_id,
	// 	usaha.Rak_id,
	// 	usaha.Version,

	if input.Rak_id != nil {
		usaha.Rak_id = input.Rak_id
	}

	if input.Version != nil {
		usaha.Version = input.Version
	}

	if input.Rak_code != nil {
		usaha.Rak_code = input.Rak_code
	}

	if input.Rak_ket != nil {
		usaha.Rak_ket = input.Rak_ket
	}
	if input.Warehouse_id != nil {
		usaha.Warehouse_id = input.Warehouse_id
	}

	err = app.models.Rak.Update(usaha)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rak": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteRakHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Rak.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Rak successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listRakHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code          string
		Warehousename string
		Ket           string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Code = app.readString(qs, "code", "")
	input.Warehousename = app.readString(qs, "warehousename", "")
	input.Ket = app.readString(qs, "ket", "")

	// input.Filters.Page = app.readInt(qs, "page", 1, v)
	// input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "rn")
	input.Filters.SortSafelist = []string{"rn", "name", "-rn", "-name"}

	// input.Filters.Sort = app.readString(qs, "sort", "rn")
	// input.Filters.SortSafelist = []string{"id", "name","-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	usahas, metadata, err := app.models.Rak.GetAll(input.Code, input.Warehousename, input.Ket, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": usahas, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
