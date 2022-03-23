package data

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/twinj/uuid"

	"greenlight.alexedwards.net/internal/validator"
)

type Stok struct {
	Qty            *float64      `json:"qty"`
	ID             *string       `json:"id"`
	CreatedAt      *time.Time    `json:"-"`
	Code           *string       `json:"produk_code"`
	Ket            *string       `json:"produk_ket"`
	Version        *int32        `json:"version"`
	Rn             *int32        `json:"rn"`
	Buy            *float64      `json:"buy"`
	Sell           *float64      `json:"sell"`
	Year           *string       `json:"year"`
	Chasis         *string       `json:"chasis"`
	BrandID        *int64        `json:"brand_id"`
	ModelID        *int64        `json:"model_id"`
	BrandName      *string       `json:"brandname"`
	ModelName      *string       `json:"modelname"`
	JsonStokDetail []*StokDetail `json:"jsonstokdetail,omitempty"`
}

type StokDetail struct {
	Id *string `json:"id,omitempty"`
	// Created_at     *time.Time `json:"-"`
	Version        *int32   `json:"version,omitempty"`
	Qty            *float64 `json:"qty,omitempty"`
	Satuan         *string  `json:"satuan,omitempty"`
	Rak_id         *int64   `json:"rak_id,omitempty"`
	Warehouse_id   *int64   `json:"warehouse_id,omitempty"`
	Stok_id        *string  `json:"stok_id,omitempty"`
	Rak_code       *string  `json:"rak_code,omitempty"`
	Name_warehouse *string  `json:"name_warehouse,omitempty"`
}

func ValidateStok(v *validator.Validator, usaha Stok) {
	v.Check(usaha.Code != nil, "Product Code", "harus diisi")
}

type StokModel struct {
	DB *sql.DB
}

func (m StokModel) Insert(usaha *Stok) error {

	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	stok_id := uuid.NewV4()
	stmtstok := (`
		INSERT INTO stok (produk_code, produk_ket,buy,sell,year,chasis,brand_id,model_id,id) 
		VALUES ($1, $2,$3,$4,$5,$6,$7,$8,$9)`)

	args := []interface{}{usaha.Code, usaha.Ket, usaha.Buy, usaha.Sell, usaha.Year, usaha.Chasis, usaha.BrandID, usaha.ModelID, stok_id}

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	_, err = tx.ExecContext(ctx, stmtstok, args...)
	if err != nil {
		str := fmt.Sprintf("%v", args)
		dataQuery := "error insert: " + stmtstok + " " + str
		log.Println(dataQuery)
		tx.Rollback()
		return err
	}

	if len(usaha.JsonStokDetail) > 0 {

		sqlStr := "insert into stok_detail(id,qty,satuan,rak_id,warehouse_id,stok_id) VALUES"
		vals := []interface{}{}

		for i, row := range usaha.JsonStokDetail {

			stok_detail_id := uuid.NewV4()

			//fmt.Println(row.Rak_code)
			//sqlStr += "($1, $2, $3),"
			sqlStr += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d),",
				i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)

			//	sqlStr += "(?),"
			vals = append(vals, stok_detail_id, row.Qty, row.Satuan, row.Rak_id, row.Warehouse_id, stok_id)
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		_, err = tx.ExecContext(ctx, sqlStr, vals...)
		if err != nil {
			str := fmt.Sprintf("%v", vals)
			dataQuery := "error insert: " + sqlStr + " " + str
			log.Println(dataQuery)
			tx.Rollback()
			return err
		}

	}
	err = tx.Commit()
	if err != nil {
		log.Printf("Failed insert Commit")
		return err

	}
	return nil
}

func (m StokModel) Get(id string) (*Stok, error) {
	if len(id) < 1 {
		return nil, ErrRecordNotFound
	}

	query := `select coalesce(g.total,0)total,a.id,a.produk_code, a.produk_ket,a.buy,a.sell,a.year,a.chasis,a.brand_id,a.model_id,
	b.name brandname,c.name modelname,a.version,
	d.qty,d.satuan,d.rak_id,d.warehouse_id,e.rak_code,f.name_warehouse
	from stok a
	left outer join brand b on b.id=a.brand_id
	left outer join brandmodel c on c.id=a.model_id
	left outer join  stok_detail d on d.stok_id=a.id
	left outer join rak e on e.rak_id=d.rak_id
	left outer join warehouse  f on f.warehouse_id=d.warehouse_id
	left outer join (select sum(qty)total,stok_id  from stok_detail group by stok_id)g on g.stok_id=a.id
	where a.id = $1`

	rows, err := m.DB.Query(query, id)
	if err != nil {
		return nil, err

	}
	defer rows.Close()

	s := Stok{}

	detail := []*StokDetail{}

	for rows.Next() {

		u := StokDetail{}
		err = rows.Scan(
			&s.Qty, //total qty detail
			&s.ID,
			&s.Code,
			&s.Ket,
			&s.Buy,
			&s.Sell,
			&s.Year,
			&s.Chasis,
			&s.BrandID,
			&s.ModelID,
			&s.BrandName,
			&s.ModelName,
			&s.Version,
			//untuk stok_detil
			&u.Qty,
			&u.Satuan,
			&u.Rak_id,
			&u.Warehouse_id,
			&u.Rak_code,
			&u.Name_warehouse,
		)

		if err != nil {
			return nil, err
		}
		detail = append(detail, &u)
	}

	if len(detail) > 0 {
		s.JsonStokDetail = detail

	} else {
		s.JsonStokDetail = nil
	}

	if err = rows.Err(); err != nil {

		return nil, err
	}

	return &s, nil
}

func (m StokModel) Update(usaha *Stok) error {

	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := (`
	update stok
	set produk_code=$1,produk_ket=$2,buy=$3,sell=$4,year=$5,chasis=$6,brand_id=$7,model_id=$8,modified_at = now(), version = version + 1
	where id=$9 and  version = $10
	RETURNING version`)

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

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		str := fmt.Sprintf("%v", args)
		dataQuery := "error update stok: " + query + " " + str
		log.Println(dataQuery)
		tx.Rollback()
		return err
	}

	querydel := (`
        DELETE FROM stok_detail
        WHERE stok_id = $1`)

	_, err = tx.ExecContext(ctx, querydel, usaha.ID)
	if err != nil {
		str := fmt.Sprintf("%v", args)
		dataQuery := "error delete all stok_detail: " + querydel + " " + str
		log.Println(dataQuery)
		tx.Rollback()
		return err
	}

	if len(usaha.JsonStokDetail) > 0 {

		sqlStr := "insert into stok_detail(id,qty,satuan,rak_id,warehouse_id,stok_id) VALUES"
		vals := []interface{}{}

		for i, row := range usaha.JsonStokDetail {

			stok_detail_id := uuid.NewV4()

			//fmt.Println(row.Rak_code)
			//sqlStr += "($1, $2, $3),"
			sqlStr += fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d),",
				i*6+1, i*6+2, i*6+3, i*6+4, i*6+5, i*6+6)

			//	sqlStr += "(?),"
			vals = append(vals, stok_detail_id, row.Qty, row.Satuan, row.Rak_id, row.Warehouse_id, usaha.ID)
		}

		//trim the last ,
		sqlStr = sqlStr[0 : len(sqlStr)-1]

		_, err = tx.ExecContext(ctx, sqlStr, vals...)
		if err != nil {
			str := fmt.Sprintf("%v", vals)
			dataQuery := "error insert update stok_detail: " + sqlStr + " " + str
			log.Println(dataQuery)
			tx.Rollback()
			return err
		}

	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed insert Commit Update stok_detail")
		return err

	}

	return nil
}

func (m StokModel) Delete(id string) error {
	if len(id) < 1 {
		return ErrRecordNotFound
	}

	ctx := context.Background()
	tx, err := m.DB.BeginTx(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	query := (`
        DELETE FROM stok_detail
        WHERE stok_id = $1`)

	_, err = tx.ExecContext(ctx, query, id)
	if err != nil {
		str := fmt.Sprintf("%v", id)
		dataQuery := "delete stok_detail: " + query + " " + str
		log.Println(dataQuery)
		tx.Rollback()
		return err
	}

	querydel := (`
	DELETE FROM stok
	WHERE id = $1`)

	_, err = tx.ExecContext(ctx, querydel, id)
	if err != nil {
		str := fmt.Sprintf("%v", id)
		dataQuery := "error delete all stok: " + querydel + " " + str
		log.Println(dataQuery)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		log.Printf("Failed delete stok dan stok_detail")
		return err

	}

	return nil
}

func (m StokModel) GetAll(code string, ket string, brandname string, modelname string, filters Filters) ([]*Stok, Metadata, error) {

	query := fmt.Sprintf(`
	select count(*) OVER(),coalesce(d.qty,0)qty,a.id,a.produk_code, a.produk_ket,a.buy,a.sell,a.year,a.chasis,a.brand_id,a.model_id,b.name brandname,c.name modelname
	from stok a
	left outer join brand b on b.id=a.brand_id
	left outer join brandmodel c on c.id=a.model_id
	left outer join (select sum(qty)qty,stok_id  from stok_detail group by stok_id)d on d.stok_id=a.id
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
		usaha := &Stok{}

		err := rows.Scan(
			&totalRecords,
			&usaha.Qty,
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

		usahas = append(usahas, usaha)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return usahas, metadata, nil
}
