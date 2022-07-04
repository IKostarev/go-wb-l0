package service

import (
	"fmt"
	"github.com/IKostarev/go-wb-l0/internal/model"
	"github.com/jmoiron/sqlx"
	"log"
)

var Db *sqlx.DB

var InitQuery = `
CREATE TABLE IF NOT EXISTS orders
	(order_uid TEXT UNIQUE,
	track_number TEXT,
	entry TEXT,
	locale TEXT,
	internal_signature TEXT,
	customer_id TEXT,
	delivery_service TEXT,
	shardkey TEXT,
	sm_id INT,
	date_created TEXT,
	oof_shard TEXT
);

CREATE TABLE IF NOT EXISTS deliveries
	(order_uid TEXT UNIQUE,
	name TEXT,
	phone TEXT,
	zip TEXT,
	city TEXT,
	address TEXT,
	region TEXT,
	email TEXT
);

CREATE TABLE IF NOT EXISTS payments
	(order_uid TEXT UNIQUE,
	request_id TEXT,
	currency TEXT,
	provider TEXT,
	amount INT,
	payment_dt INT,
	bank TEXT,
	delivery_cost INT,
	goods_total INT,
	custom_fee INT
);

CREATE TABLE IF NOT EXISTS items
	(order_uid TEXT,
	chrt_id INT,
	track_number TEXT,
	price TEXT,
	rid TEXT,
	name TEXT,
	sale INT,
	size TEXT,
	total_price INT,
	nm_id INT,
	brand TEXT,
	status INT
);

CREATE TABLE IF NOT EXISTS nats
	(
	    host TEXT UNIQUE,
	    last TEXT
);`

func DBInit(cfg model.Config) *sqlx.DB {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s", cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))

	if err != nil {
		log.Fatalf("Error connect db: %s", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error ping db: %s", err)
	}

	_, err = db.Exec(InitQuery)

	return db
}
