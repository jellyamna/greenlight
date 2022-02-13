package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type Warehouse struct {
	Perusahaan_Id     int64     `json:"perusahaan_id"`
	Warehouse_id      int64     `json:"warehouse_id"`
	Name_perusahaan   string    `json:"name_perusahaan"`
	Name_warehouse    string    `json:"name_warehouse"`
	Address_warehouse string    `json:"address_warehouse"`
	Tlp_warehouse     string    `json:"tlp_warehouse"`
	Ket_warehouse     string    `json:"ket_warehouse"`
	User_modified     string    `json:"user_modified"`
	Created_at        time.Time `json:"created_at"`
	Version           int32     `json:"Version"`
}

func ValidateWarehouse(v *validator.Validator, usaha *Warehouse) {
	v.Check(usaha.Name_warehouse != "", "Nama", "harus diisi")
	v.Check(len(usaha.Address_warehouse) >= 1888, "Alamat", "harus lebih besar dari 10 karakter")

	v.Check(usaha.Tlp_warehouse != "", "Tlp", "harus  diisi")

}

type WarehouseModel struct {
	DB *sql.DB
}

func (m WarehouseModel) Insert(usaha *Warehouse) error {

	query := `
		INSERT INTO warehouse (name_warehouse, address_warehouse, tlp_warehouse, ket_warehouse,user_modified,perusahaan_id) 
		VALUES ($1, $2, $3, $4,$5,$6)
		RETURNING warehouse_id, created_at, version`

	// query := `
	// 	INSERT INTO perusahaan (name, address, tlp, npwp,rek,ket)
	// 	VALUES ($1, $2, $3, $4,$5,$6)
	//     RETURNING id, created_at, version`

	args := []interface{}{usaha.Name_warehouse, usaha.Address_warehouse, usaha.Tlp_warehouse, usaha.Ket_warehouse, usaha.User_modified, usaha.Perusahaan_Id}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.Warehouse_id, &usaha.Created_at,
		&usaha.Version)
}

func (m WarehouseModel) Get(id int64) (*Warehouse, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	SELECT a.perusahaan_id,a.warehouse_id,b.name name_perusahaan,a.name_warehouse, a.address_warehouse, a.tlp_warehouse, a.ket_warehouse,a.user_modified,
	a.created_at,a.version
	FROM warehouse a
	inner join perusahaan b on a.perusahaan_id=b.id
	WHERE a.warehouse_id= $1 `

	var usaha Warehouse

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.Perusahaan_Id,
		&usaha.Warehouse_id,
		&usaha.Name_perusahaan,
		&usaha.Name_warehouse,
		&usaha.Address_warehouse,
		&usaha.Tlp_warehouse,
		&usaha.Ket_warehouse,
		&usaha.User_modified,
		&usaha.Created_at,
		&usaha.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &usaha, nil
}

func (m WarehouseModel) Update(usaha *Warehouse) error {
	query := `
	UPDATE warehouse 
	SET name_warehouse = $1, address_warehouse = $2, tlp_warehouse = $3, ket_warehouse = $4, version = version + 1, modified_at= now(),perusahaan_id = $5
	WHERE warehouse_id = $6 AND version = $7
	RETURNING version`

	args := []interface{}{
		usaha.Name_warehouse,
		usaha.Address_warehouse,
		usaha.Tlp_warehouse,
		usaha.Ket_warehouse,
		usaha.Perusahaan_Id,
		usaha.Warehouse_id,
		usaha.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m WarehouseModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM warehouse
        WHERE warehouse_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m WarehouseModel) GetAll(name string, alamat string, filters Filters) ([]*Warehouse, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(),a.perusahaan_id,a.warehouse_id,b.name name_perusahaan,a.name_warehouse, a.address_warehouse, a.tlp_warehouse, a.ket_warehouse,a.user_modified,
	a.created_at,a.version
	FROM warehouse a
	inner join perusahaan b on a.perusahaan_id=b.id
	where lower(a.name_warehouse) like lower('%%` + name + `%%') and  lower(a.address_warehouse) like lower('%%` + alamat + `%%') 
	ORDER BY a.created_at
	LIMIT $1 OFFSET $2`)

	// query := fmt.Sprintf(`
	// SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version,0
	// FROM perusahaan
	// WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// ORDER BY %s %s
	// LIMIT $2 OFFSET $3`, "name", "asc")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//args := []interface{}{name}
	args := []interface{}{filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println("err-db")
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	usahas := []*Warehouse{}

	for rows.Next() {
		var usaha Warehouse

		// SELECT count(*) OVER(),a.perusahaan_id,a.warehouse_id,b.name name_perusahaan,a.name_warehouse,
		// a.address_warehouse, a.tlp_warehouse, a.ket_warehouse,a.user_modified,
		// a.created_at,a.version
		err := rows.Scan(
			&totalRecords,
			&usaha.Perusahaan_Id,
			&usaha.Warehouse_id,
			&usaha.Name_perusahaan,
			&usaha.Name_warehouse,
			&usaha.Address_warehouse,
			&usaha.Tlp_warehouse,
			&usaha.Ket_warehouse,
			&usaha.User_modified,
			&usaha.Created_at,
			&usaha.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		usahas = append(usahas, &usaha)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return usahas, metadata, nil
}
