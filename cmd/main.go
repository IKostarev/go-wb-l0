package main

import (
	"github.com/IKostarev/go-wb-l0/internal/model"
	"github.com/IKostarev/go-wb-l0/pkg/service"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

var dbConf = model.Config{
	Host:     "localhost",
	Port:     "5432",
	Username: "test",
	Password: "test",
	DBName:   "wb_test",
	SSLMode:  "disable",
}

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)

	db = service.DBInit(dbConf)
	db.SetMaxOpenConns(100)
	service.Db = db

}
