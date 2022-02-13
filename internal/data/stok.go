package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type Stok struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Code      string    `json:"produk_code"`
	Ket       string    `json:"produk_ket"`
	Version   int32     `json:"version"`
	Rn        int32     `json:"rn"`
	Buy       float64   `json:"buy"`
	Sell      float64   `json:"sell"`
	Year      string    `json:"year"`
	Chasis    string    `json:"chasis"`
	BrandID   int64     `json:"brand_id"`
	ModelID   int64     `json:"model_id"`
	BrandName string    `json:"brandname"`
	ModelName string    `json:"modelname"`
}

func ValidateStok(v *validator.Validator, usaha *Stok) {
	v.Check(usaha.Code != "", "Product Code", "harus diisi")
}

type StokModel struct {
	DB *sql.DB
}

func (m StokModel) Insert(usaha *Stok) error {
	query := `
		INSERT INTO stok (produk_code, produk_ket,buy,sell,year,chasis,brand_id,model_id) 
		VALUES ($1, $2,$3,$4,$5,$6,$7,$8)
        RETURNING id, created_at, version`

	args := []interface{}{usaha.Code, usaha.Ket, usaha.Buy, usaha.Sell, usaha.Year, usaha.BrandID, usaha.ModelID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.ID, &usaha.CreatedAt,
		&usaha.Version)
}

func (m StokModel) Get(id int64) (*Stok, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
	select a.id,a.produk_code, a.produk_ket,a.buy,a.sell,a.year,a.chasis,a.brand_id,a.model_id,b.name brandname,c.name modelname
	from stok a
	left outer join brand b on b.id=a.brand_id
	left outer join brandmodel c on c.id=a.model_id
	where a.id = $1`

	var usaha Stok

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.ID,
		&usaha.Code,
		&usaha.Ket,
		&usaha.Buy,
		&usaha.Sell,
		&usaha.Year,
		&usaha.Chasis,
		&usaha.BrandID,
		&usaha.ModelID,
		&usaha.BrandName,
		&usaha.ModelName,
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

func (m StokModel) Update(usaha *Stok) error {
	query := `
	update stok
	set produk_code=$1,produk_ket=$2,buy=$3,sell=$4,year=$5,chasis=$6,brand_id=$7,model_id=$8,modified_at = now(), version = version + 1
	where id=$9 and  version = $10
	RETURNING version`

	args := []interface{}{
		usaha.Code,
		usaha.Ket,
		usaha.Buy,
		usaha.Sell,
		usaha.Year,
		usaha.Chasis,
		usaha.BrandID,
		usaha.ModelID,
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

func (m StokModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM stok
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

func (m StokModel) GetAll(code string, ket string, brandname string, modelname string, filters Filters) ([]*Stok, Metadata, error) {

	query := fmt.Sprintf(`
	select count(*) OVER(),a.id,a.produk_code, a.produk_ket,a.buy,a.sell,a.year,a.chasis,a.brand_id,a.model_id,b.name brandname,c.name modelname
	from stok a
	left outer join brand b on b.id=a.brand_id
	left outer join brandmodel c on c.id=a.model_id
	where lower(a.produk_code) like lower('%%` + code + `%%')  and lower(a.produk_ket) like lower('%%` + ket + `%%')
	and  lower(b.name) like lower('%%` + brandname + `%%')
	and  lower(c.name) like lower('%%` + modelname + `%%')
	ORDER BY a.created_at desc
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
	usahas := []*Stok{}

	for rows.Next() {
		var usaha Stok

		err := rows.Scan(
			&totalRecords,
			&usaha.ID,
			&usaha.Code,
			&usaha.Ket,
			&usaha.Buy,
			&usaha.Sell,
			&usaha.Year,
			&usaha.Chasis,
			&usaha.BrandID,
			&usaha.ModelID,
			&usaha.BrandName,
			&usaha.ModelName,
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
