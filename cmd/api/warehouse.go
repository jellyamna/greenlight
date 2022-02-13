package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"greenlight.alexedwards.net/internal/data"
	"greenlight.alexedwards.net/internal/validator"
)

func (app *application) createWarehouseHandler(w http.ResponseWriter, r *http.Request) {

	var input struct {
		Perusahaan_Id        		int64     	`json:"perusahaan_id"`
		Name_warehouse        		string     	`json:"name_warehouse"`
		Address_warehouse        	string     	`json:"address_warehouse"`
		Tlp_warehouse        		string     	`json:"tlp_warehouse"`
		Ket_warehouse        		string     	`json:"ket_warehouse"`
		User_modified        		string     	`json:"user_modified"`
	}


	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}


	usaha := &data.Warehouse{
		Perusahaan_Id:   	input.Perusahaan_Id,
		Name_warehouse:    	input.Name_warehouse,
		Address_warehouse: 	input.Address_warehouse,
		Tlp_warehouse:  	input.Tlp_warehouse,
		Ket_warehouse:  	input.Ket_warehouse,
		User_modified:  	input.User_modified,
	}

	// v := validator.New()

	// // if data.ValidatePerusahaan(v, usaha); !v.Valid() {
	// // 	app.failedValidationResponse(w, r, v.Errors)
	// // 	return
	// // }

	err = app.models.Warehouse.Insert(usaha)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/warehouse/%d", usaha.Warehouse_id))

	err = app.writeJSON(w, http.StatusCreated, envelope{"usaha": usaha}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Warehouse.Get(id)
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

func (app *application) updateWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	usaha, err := app.models.Warehouse.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}


	// var input struct {
	// 	Name   		*string     `json:"name"`
	// 	Address 	*string     `json:"address"`
	// 	Tlp   		*string   	`json:"tlp"`
	// 	Npwp   		*string  	`json:"npwp"`
	// 	Rek   		*string     `json:"rek"`
	// 	Ket 		*string 	`json:"ket"`
	// }

	var input struct {
		Perusahaan_Id        		*int64     	`json:"perusahaan_id"`
		Warehouse_id        		*int64     	`json:"warehouse_id"`
		Name_perusahaan        		*string     `json:"name_perusahaan"`
		Name_warehouse        		*string     `json:"name_warehouse"`
		Address_warehouse        	*string     `json:"address_warehouse"`
		Tlp_warehouse        		*string     `json:"tlp_warehouse"`
		Ket_warehouse        		*string     `json:"ket_warehouse"`
		User_modified        		*string     `json:"user_modified"`
		Created_at        			*time.Time  `json:"created_at"`
		Version        				*int32     	`json:"Version"`
	}

	

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Perusahaan_Id != nil {
		usaha.Perusahaan_Id = *input.Perusahaan_Id
	}

	//Perusahaan_Id        		*int64     	`json:"perusahaan_id"`
		// Warehouse_id        		*int64     	`json:"warehouse_id"`
		// Name_perusahaan        		*string     `json:"name_perusahaan"`
		// Name_warehouse        		*string     `json:"name_warehouse"`
		// Address_warehouse        	*string     `json:"address_warehouse"`
		// Tlp_warehouse        		*string     `json:"tlp_warehouse"`
		// Ket_warehouse        		*string     `json:"ket_warehouse"`
		// User_modified        		*string     `json:"user_modified"`
		// Created_at        			*time.Time  `json:"created_at"`
		// Version        				*int32     	`json:"Version"`

	if input.Warehouse_id != nil {
		usaha.Warehouse_id = *input.Warehouse_id
	}
	if input.Name_warehouse != nil {
		usaha.Name_warehouse = *input.Name_warehouse
	}
	if input.Address_warehouse != nil {
		usaha.Address_warehouse = *input.Address_warehouse
	}

	if input.Tlp_warehouse != nil {
		usaha.Tlp_warehouse = *input.Tlp_warehouse
	}


	if input.Ket_warehouse != nil {
		usaha.Ket_warehouse = *input.Ket_warehouse
	}



	// v := validator.New()

	// if data.ValidatePerusahaan(v, usaha); !v.Valid() {
	// 	app.failedValidationResponse(w, r, v.Errors)
	// 	return
	// }

	err = app.models.Warehouse.Update(usaha)
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

func (app *application) deleteWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Warehouse.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "Warehouse successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listWarehouseHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string
		Alamat string
		data.Filters
	}

	v := validator.New()

	qs := r.URL.Query()

	input.Name = app.readString(qs, "name", "")
	input.Alamat = app.readString(qs, "alamat", "")
	

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

	usahas, metadata, err := app.models.Warehouse.GetAll(input.Name,input.Alamat, input.Filters)
	if err != nil {
		fmt.Println("sampai-err")
		app.serverErrorResponse(w, r, err)
		return
	}

	
	err = app.writeJSON(w, http.StatusOK, envelope{"warehouse": usahas, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
