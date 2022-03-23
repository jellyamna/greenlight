package main

import (
	"errors"
	"net/http"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

// var wg sync.WaitGroup

func (app *application) createStokHandler(w http.ResponseWriter, r *http.Request) {

	// type input struct {
	// 	Rak_code     string `json:"rak_code"`
	// 	Rak_ket      string `json:"rak_ket"`
	// 	Warehouse_id int32  `json:"warehouse_id"`
	// }

	stok := data.Stok{}

	err := app.readJSONArray(w, r, &stok)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	err = app.models.Stok.Insert(&stok)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

}

func (app *application) showStokHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParamString(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Stok.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"stok": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateStokHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParamString(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Stok.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	input := data.Stok{}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Code != nil {
		usaha.Code = input.Code
	}

	if input.Ket != nil {
		usaha.Ket = input.Ket
	}

	if input.Buy != nil {
		usaha.Buy = input.Buy
	}

	if input.Sell != nil {
		usaha.Sell = input.Sell
	}

	if input.Year != nil {
		usaha.Year = input.Year
	}

	if input.Chasis != nil {
		usaha.Chasis = input.Chasis
	}

	if input.BrandID != nil {
		usaha.BrandID = input.BrandID
	}

	if input.ModelID != nil {
		usaha.ModelID = input.ModelID
	}

	if input.Version != nil {
		usaha.Version = input.Version
	}

	if len(input.JsonStokDetail) > 0 {
		usaha.JsonStokDetail = nil
		usaha.JsonStokDetail = input.JsonStokDetail
	} else {
		usaha.JsonStokDetail = nil
	}

	err = app.models.Stok.Update(usaha)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"stok": usaha}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteStokHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParamString(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Stok.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Stok successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listStokHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code      string
		Ket       string
		Brandname string
		Modelname string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Code = app.readString(qs, "code", "")
	input.Ket = app.readString(qs, "ket", "")
	input.Brandname = app.readString(qs, "brandname", "")
	input.Modelname = app.readString(qs, "modelname", "")

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

	usahas, metadata, err := app.models.Stok.GetAll(input.Code, input.Ket, input.Brandname, input.Modelname, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": usahas, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
