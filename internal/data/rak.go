package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"greenlight.alexedwards.net/internal/validator"
)

type Rak struct {
	Rak_id         *int64     `json:"rak_id"`
	Created_at     *time.Time `json:"created_at"`
	Rak_code       *string    `json:"rak_code"`
	Rak_ket        *string    `json:"rak_ket"`
	Version        *int32     `json:"version"`
	User_modified  *string    `json:"user_modified"`
	Warehouse_id   *int32     `json:"warehouse_id"`
	Name_warehouse *string    `json:"Name_warehouse"`
}

func ValidateRak(v *validator.Validator, rak *Rak) {
	v.Check(rak.Rak_code != nil, "Code Rak", "harus diisi")

}

type RakModel struct {
	DB *sql.DB
}

func (m RakModel) Insert(usaha *[]RakMultiInsert) error {

	sqlStr := "INSERT INTO rak (rak_code,rak_ket,warehouse_id) VALUES "
	vals := []interface{}{}
	//valueStrings := make([]string, 0, 1)

	for i, row := range *usaha {

		fmt.Println(row.Rak_code)
		//sqlStr += "($1, $2, $3),"
		sqlStr += fmt.Sprintf("($%d,$%d,$%d),",
			i*3+1, i*3+2, i*3+3)

		//	sqlStr += "(?),"
		vals = append(vals, row.Rak_code, row.Rak_ket, row.Warehouse_id)
	}

	//trim the last ,
	sqlStr = sqlStr[0 : len(sqlStr)-1]
	//prepare the statement
	stmt, _ := m.DB.Prepare(sqlStr)

	//format all vals at once
	_, err := stmt.Exec(vals...)
	if err != nil {
		return err
	}
	return nil

	// query := `INSERT INTO rak (rak_code, rak_ket, warehouse_id)
	// VALUES ($1, $2, $3)
	// RETURNING rak_id, created_at, version`

	// args := []interface{}{usaha.Rak_code, usaha.Rak_ket, usaha.Warehouse_id}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	// return m.DB.QueryRowContext(ctx, query, args...).Scan(&usaha.Rak_id, &usaha.Created_at,
	// 	&usaha.Version)
}

func (m RakModel) Get(id int64) (*Rak, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := ` select a.rak_id,a.created_at,a.rak_code,a.rak_ket,a.version,a.user_modified,a.warehouse_id,b.name_warehouse
	from rak a
	inner join warehouse b on a.warehouse_id=b.warehouse_id
	where a.rak_id=$1 `

	var usaha Rak

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	//a.rak_id,a.created_at,a.rak_code,a.rak_ket,a.version,
	//a.modified_at,a.user_modified,a.warehouse_id,b.name_warehouse

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&usaha.Rak_id,
		&usaha.Created_at,
		&usaha.Rak_code,
		&usaha.Rak_ket,
		&usaha.Version,
		&usaha.User_modified,
		&usaha.Warehouse_id,
		&usaha.Name_warehouse,
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

func (m RakModel) Update(usaha *Rak) error {
	query := `
	UPDATE rak 
	SET rak_code = $1, rak_ket = $2, version = version + 1, modified_at= now(),warehouse_id = $3
	WHERE rak_id = $4 AND version = $5
	RETURNING version`

	args := []interface{}{
		usaha.Rak_code,
		usaha.Rak_ket,
		usaha.Warehouse_id,
		usaha.Rak_id,
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

func (m RakModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM rak
        WHERE rak_id = $1`

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

func (m RakModel) GetAll(code string, warehousename string, ket string, filters Filters) ([]*Rak, Metadata, error) {

	query := fmt.Sprintf(`
	SELECT count(*) OVER(),a.rak_id,a.created_at,a.rak_code,a.rak_ket,a.version,
	a.user_modified,a.warehouse_id,b.name_warehouse
	FROM rak a
	inner join warehouse b on a.warehouse_id=b.warehouse_id
	where lower(a.rak_code) like lower('%%` + code + `%%') and  lower(b.name_warehouse) like lower('%%` + warehousename + `%%') 
	and lower(a.rak_ket) like lower('%%` + ket + `%%')
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
	usahas := []*Rak{}

	for rows.Next() {
		var usaha Rak
		err := rows.Scan(
			&totalRecords,
			&usaha.Rak_id,
			&usaha.Created_at,
			&usaha.Rak_code,
			&usaha.Rak_ket,
			&usaha.Version,
			&usaha.User_modified,
			&usaha.Warehouse_id,
			&usaha.Name_warehouse,
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
