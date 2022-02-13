package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

func (app *application) createBrandHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ID        int64     `json:"id"`
		CreatedAt time.Time `json:"-"`
		Name      string    `json:"name"`
		Ket       string    `json:"ket"`
		Version   int32     `json:"version"`
		Rn        int32     `json:"rn"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	usaha := &data.Brand{
		ID:      input.ID,
		Name:    input.Name,
		Ket:     input.Ket,
		Version: input.Version,
	}

	err = app.models.Brand.Insert(usaha)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/brand/%d", usaha.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"usaha": usaha}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Brand.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"usaha": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Brand.Get(id)
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
		ID        *int64     `json:"id"`
		CreatedAt *time.Time `json:"-"`
		Name      *string    `json:"name"`
		Ket       *string    `json:"ket"`
		Version   *int32     `json:"version"`
		Rn        *int32     `json:"rn"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.ID != nil {
		usaha.ID = *input.ID
	}

	if input.Name != nil {
		usaha.Name = *input.Name
	}

	if input.Ket != nil {
		usaha.Ket = *input.Ket
	}

	err = app.models.Brand.Update(usaha)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"usaha": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteBrandHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Brand.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Brand successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listBrandHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name string
		Ket  string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Ket = app.readString(qs, "ket", "")

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

	usahas, metadata, err := app.models.Brand.GetAll(input.Name, input.Ket, input.Filters)
	if err != nil {
		fmt.Println("sampai-err")
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": usahas, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
