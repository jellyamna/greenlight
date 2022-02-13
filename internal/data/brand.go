package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type Brand struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Ket       string    `json:"ket"`
	Version   int32     `json:"version"`
	Rn        int32     `json:"rn"`
}

func ValidateBrand(v *validator.Validator, usaha *Brand) {
	v.Check(usaha.Name != "", "Nama", "harus diisi")
}

type BrandModel struct {
	DB *sql.DB
}

func (m BrandModel) Insert(usaha *Brand) error {
	query := `
		INSERT INTO brand (name, ket) 
		VALUES ($1, $2)
        RETURNING id, created_at, version`

	args := []interface{}{usaha.Name, usaha.Ket}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.ID, &usaha.CreatedAt,
		&usaha.Version)
}

func (m BrandModel) Get(id int64) (*Brand, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id,created_at,name,ket,version
        FROM brand
        WHERE id = $1`

	var usaha Brand

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.ID,
		&usaha.CreatedAt,
		&usaha.Name,
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

func (m BrandModel) Update(usaha *Brand) error {
	query := `
	UPDATE brand 
	SET name = $1, ket = $2, modified_at = now(), version = version + 1
	WHERE id = $3 AND version = $4
	RETURNING version`

	args := []interface{}{
		usaha.Name,
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

func (m BrandModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM brand
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

func (m BrandModel) GetAll(name string, ket string, filters Filters) ([]*Brand, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(),id,created_at,name,ket,version
	FROM brand 
	where lower(name) like lower('%%` + name + `%%')  and lower(ket) like lower('%%` + ket + `%%')
	ORDER BY created_at desc
	LIMIT $1 OFFSET $2`)

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
	usahas := []*Brand{}

	for rows.Next() {
		var usaha Brand

		//SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version-10
		err := rows.Scan(
			&totalRecords,
			&usaha.ID,
			&usaha.CreatedAt,
			&usaha.Name,
			&usaha.Ket,
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
