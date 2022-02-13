package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Warehouse       WarehouseModel
	Perusahaans     PerusahaanModel
	Movies          MovieModel
	Permissions     PermissionModel
	Tokens          TokenModel
	Users           UserModel
	Rak             RakModel
	Brand           BrandModel
	BrandAssetModel BrandAssetModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Warehouse:       WarehouseModel{DB: db},
		Perusahaans:     PerusahaanModel{DB: db},
		Movies:          MovieModel{DB: db},
		Permissions:     PermissionModel{DB: db},
		Tokens:          TokenModel{DB: db},
		Users:           UserModel{DB: db},
		Rak:             RakModel{DB: db},
		Brand:           BrandModel{DB: db},
		BrandAssetModel: BrandAssetModel{DB: db},
	}
}
