package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type Perusahaan struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Address   string    `json:"address"`
	Tlp       string    `json:"tlp"`
	Npwp      string    `json:"npwp"`
	Rek       string    `json:"rek"`
	Ket       string    `json:"ket"`
	Version   int32     `json:"version"`
	Rn        int32     `json:"rn"`
}

func ValidatePerusahaan(v *validator.Validator, usaha *Perusahaan) {
	v.Check(usaha.Name != "", "Nama", "harus diisi")
	v.Check(len(usaha.Address) >= 1888, "Alamat", "harus lebih besar dari 10 karakter")

	v.Check(usaha.Tlp != "", "Tlp", "harus  diisi")
	v.Check(usaha.Npwp != "", "NPWP", "harus  diisi")
}

type PerusahaanModel struct {
	DB *sql.DB
}

func (m PerusahaanModel) Insert(usaha *Perusahaan) error {
	query := `
		INSERT INTO perusahaan (name, address, tlp, npwp,rek,ket) 
		VALUES ($1, $2, $3, $4,$5,$6)
        RETURNING id, created_at, version`

	args := []interface{}{usaha.Name, usaha.Address, usaha.Tlp, usaha.Npwp, usaha.Rek, usaha.Ket}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.ID, &usaha.CreatedAt,
		&usaha.Version)
}

func (m PerusahaanModel) Get(id int64) (*Perusahaan, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, created_at,name, address, tlp, npwp,rek,ket,version
        FROM perusahaan
        WHERE id = $1`

	var usaha Perusahaan

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.ID,
		&usaha.CreatedAt,
		&usaha.Name,
		&usaha.Address,
		&usaha.Tlp,
		&usaha.Npwp,
		&usaha.Rek,
		&usaha.Ket,
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

func (m PerusahaanModel) Update(usaha *Perusahaan) error {
	query := `
	UPDATE perusahaan 
	SET name = $1, address = $2, tlp = $3, npwp = $4, rek=$5,ket=$6, version = version + 1
	WHERE id = $7 AND version = $8
	RETURNING version`

	args := []interface{}{
		usaha.Name,
		usaha.Address,
		usaha.Tlp,
		usaha.Npwp,
		usaha.Rek,
		usaha.Ket,
		usaha.ID,
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

func (m PerusahaanModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM perusahaan
        WHERE id = $1`

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

func (m PerusahaanModel) GetAll(name string, filters Filters) ([]*Perusahaan, Metadata, error) {

	// query := " select *,row_number() over() rn from (SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version
	// FROM perusahaan
	// WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// ORDER BY name ASC) a
	// LIMIT $2 OFFSET $3 "

	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version,0
        FROM perusahaan
        WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')  
        ORDER BY %s %s 
		LIMIT $2 OFFSET $3`, "name", "asc")

	// query := fmt.Sprintf(`
	// SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version,
	// row_number() over ()rn
	// FROM perusahaan
	// WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
	// ORDER BY %s %s, id ASC
	// LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//args := []interface{}{name}
	args := []interface{}{name, filters.limit(), filters.offset()}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		fmt.Println("err-db")
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	usahas := []*Perusahaan{}

	for rows.Next() {
		var usaha Perusahaan

		//SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version-10
		err := rows.Scan(
			&totalRecords,
			&usaha.ID,
			&usaha.CreatedAt,
			&usaha.Name,
			&usaha.Address,
			&usaha.Tlp,
			&usaha.Npwp,
			&usaha.Rek,
			&usaha.Ket,
			&usaha.Version,
			&usaha.Rn,
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
