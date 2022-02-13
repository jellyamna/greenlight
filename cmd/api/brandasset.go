package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

func (app *application) createBrandAssetHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		ID        int64     `json:"id"`
		CreatedAt time.Time `json:"-"`
		Name      string    `json:"name"`
		Ket       string    `json:"ket"`
		BrandID   int64     `json:"brand_id"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	usaha := &data.BrandAsset{
		ID:      input.ID,
		Name:    input.Name,
		Ket:     input.Ket,
		BrandID: input.BrandID,
	}

	err = app.models.BrandAssetModel.Insert(usaha)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/brandasset/%d", usaha.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"usaha": usaha}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showBrandAsssetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.BrandAssetModel.Get(id)
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

func (app *application) updateBrandAssetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.BrandAssetModel.Get(id)
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
		ID        *int64    `json:"id"`
		CreatedAt time.Time `json:"-"`
		Name      *string   `json:"name"`
		Ket       *string   `json:"ket"`
		BrandID   *int64    `json:"brand_id"`
		Version   *int32    `json:"version"`
		Rn        *int32    `json:"rn"`
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

	if input.BrandID != nil {
		usaha.BrandID = *input.BrandID
	}

	err = app.models.BrandAssetModel.Update(usaha)
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

func (app *application) deleteBrandAssetHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.BrandAssetModel.Delete(id)
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

func (app *application) listBrandAssetHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name      string
		Ket       string
		BrandName string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Ket = app.readString(qs, "ket", "")
	input.BrandName = app.readString(qs, "brandname", "")

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

	usahas, metadata, err := app.models.BrandAssetModel.GetAll(input.Name, input.BrandName, input.Ket, input.Filters)
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
