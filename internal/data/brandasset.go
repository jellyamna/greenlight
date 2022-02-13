package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type BrandAsset struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Ket       string    `json:"ket"`
	Version   int32     `json:"version"`
	Rn        int32     `json:"rn"`
	BrandID   int64     `json:"brand_id"`
	BrandName string    `json:"brandname"`
}

func ValidateBrandAsset(v *validator.Validator, usaha *BrandAsset) {
	v.Check(usaha.Name != "", "Nama", "harus diisi")
	v.Check(usaha.BrandID != 0, "Brand Name", "harus diisi")
}

type BrandAssetModel struct {
	DB *sql.DB
}

func (m BrandAssetModel) Insert(usaha *BrandAsset) error {
	query := `
		insert into brandmodel(name,ket,brand_id) 
		VALUES ($1, $2,$3)
        RETURNING id, created_at, version`

	args := []interface{}{usaha.Name, usaha.Ket, usaha.BrandID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.ID, &usaha.CreatedAt,
		&usaha.Version)
}

func (m BrandAssetModel) Get(id int64) (*BrandAsset, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT b.id,b.created_at,b.name,b.ket,b.version,b.brand_id,a.name namebrand
	FROM brand  a
	inner join brandmodel b on a.id=b.brand_id
	WHERE b.id = $1`

	var usaha BrandAsset

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.ID,
		&usaha.CreatedAt,
		&usaha.Name,
		&usaha.Ket,
		&usaha.Version,
		&usaha.BrandID,
		&usaha.BrandName,
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

func (m BrandAssetModel) Update(usaha *BrandAsset) error {
	query := `
	UPDATE brandmodel 
	SET name = $1, ket = $2, modified_at = now(), version = version + 1,brand_id=$3
	WHERE id = $4 AND version = $5
	RETURNING version`

	args := []interface{}{
		usaha.Name,
		usaha.Ket,
		usaha.BrandID,
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

func (m BrandAssetModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM brandmodel
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

func (m BrandAssetModel) GetAll(name string, namebrand string, ket string, filters Filters) ([]*BrandAsset, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(),b.id,b.created_at,b.name,b.ket,b.version,b.brand_id,a.name namebrand
	FROM brand  a
	inner join brandmodel b on a.id=b.brand_id
	where lower(b.name) like lower('%%` + name + `%%') and lower(b.ket) like lower('%%` + ket + `%%') and 
	lower(a.name) like lower('%%` + namebrand + `%%')
	ORDER BY b.created_at desc
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
	usahas := []*BrandAsset{}

	for rows.Next() {
		var usaha BrandAsset

		//SELECT count(*) OVER(), id, created_at, name, address, tlp, npwp, rek,ket,version-10
		err := rows.Scan(
			&totalRecords,
			&usaha.ID,
			&usaha.CreatedAt,
			&usaha.Name,
			&usaha.Ket,
			&usaha.Version,
			&usaha.BrandID,
			&usaha.BrandName,
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
