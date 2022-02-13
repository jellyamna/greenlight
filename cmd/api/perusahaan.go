package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

func (app *application) createPerusahaanHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name   		string      `json:"name"`
		Address 	string      `json:"address"`
		Tlp   		string   	`json:"tlp"`
		Npwp   		string  	`json:"npwp"`
		Rek   		string     	`json:"rek"`
		Ket 		string 		`json:"ket"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	usaha := &data.Perusahaan{
		Name:   input.Name,
		Address:    input.Address,
		Tlp: input.Tlp,
		Npwp:  input.Npwp,
		Rek:  input.Rek,
		Ket:  input.Ket,
	}

	// v := validator.New()

	// // if data.ValidatePerusahaan(v, usaha); !v.Valid() {
	// // 	app.failedValidationResponse(w, r, v.Errors)
	// // 	return
	// // }

	err = app.models.Perusahaans.Insert(usaha)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/perusahaans/%d", usaha.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"usaha": usaha}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showPerusahaanHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Perusahaans.Get(id)
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

func (app *application) updatePerusahaanHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Perusahaans.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	
	fmt.Println("update jalan")

	var input struct {
		Name   		*string     `json:"name"`
		Address 	*string     `json:"address"`
		Tlp   		*string   	`json:"tlp"`
		Npwp   		*string  	`json:"npwp"`
		Rek   		*string     `json:"rek"`
		Ket 		*string 	`json:"ket"`
	}


	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Name != nil {
		usaha.Name = *input.Name
	}

	if input.Address != nil {
		usaha.Address = *input.Address
	}
	if input.Tlp != nil {
		usaha.Tlp = *input.Tlp
	}
	if input.Npwp != nil {
		usaha.Npwp = *input.Npwp
	}
	if input.Npwp != nil {
		usaha.Npwp = *input.Npwp
	}

	if input.Ket != nil {
		usaha.Ket = *input.Ket
	}



	// v := validator.New()

	// if data.ValidatePerusahaan(v, usaha); !v.Valid() {
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = app.models.Perusahaans.Update(usaha)
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

func (app *application) deletePerusahaanHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Perusahaans.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Perusahaan successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listPerusahaanHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	

	// input.Filters.Page = app.readInt(qs, "page", 1, v)
	// input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	input.Filters.Sort = app.readString(qs, "sort", "rn")
	input.Filters.SortSafelist = []string{"rn", "name","-rn", "-name"}

	// input.Filters.Sort = app.readString(qs, "sort", "rn")
	// input.Filters.SortSafelist = []string{"id", "name","-id", "-name"}

	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	usahas, metadata, err := app.models.Perusahaans.GetAll(input.Name, input.Filters)
	if err != nil {
		fmt.Println("sampai-err")
		app.serverErrorResponse(w, r, err)
		return
	}

	
	err = app.writeJSON(w, http.StatusOK, envelope{"Perusahaan": usahas, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
