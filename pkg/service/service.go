package service

import (
	"fmt"
	"github.com/IKostarev/go-wb-l0/cache"
	"github.com/IKostarev/go-wb-l0/internal/model"
	ctx "github.com/gorilla/context"
	"github.com/jmoiron/sqlx"
	"log"
	"net/http"
	"time"
)

var (
	Db    *sqlx.DB
	Cache *cache.Cache
)

var InitQuery = `
CREATE TABLE IF NOT EXISTS models
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

func ModelAdd(m model.Model) error {
	insertOrder := `INSERT INTO models 
	VALUES (:order_uid, :track_number, :entry,
	:locale, :internal_signature, :customer_id, :delivery_service,:shardkey,
	:sm_id, :date_created, :oof_shard)`

	insertDelivery := `INSERT INTO deliveries 
	VALUES (:order_uid, :name, :phone, :zip, :city, :address, :region, :email)`

	insertPayment := `INSERT INTO payments 
	VALUES (:order_uid, :request_id, :currency, :provider, :amount,
	:payment_dt, :bank, :delivery_cost, :goods_total, :custom_fee)`

	insertItem := `INSERT INTO items 
	VALUES (:order_uid, :chrt_id, :track_number, :price, :rid,
	:name, :sale, :size, :total_price, :nm_id, :brand, :status)`

	_, err := Db.Exec(insertOrder, m)
	if err != nil {
		log.Fatalf("При добавлении %s в таблицу models произошла ошибка:%s", m.OrderUid, err)
	}

	_, err = Db.Exec(insertDelivery, m)
	if err != nil {
		ModelRemove(m.OrderUid)
		log.Fatalf("При добавлении %s в таблицу models произошла ошибка:%s", m.OrderUid, err)
	}

	_, err = Db.Exec(insertPayment, m)
	if err != nil {
		ModelRemove(m.OrderUid)
		log.Fatalf("При добавлении %s в таблицу models произошла ошибка:%s", m.OrderUid, err)
	}

	for _, item := range m.Items {
		_, err = Db.Exec(insertItem, item)
		if err != nil {
			ModelRemove(m.OrderUid)
			log.Fatalf("При добавлении %s в таблицу models произошла ошибка:%s", m.OrderUid, err)
		}
	}

	Cache.Append(m)

	return nil
}

func ModelGet(r *http.Request, orderUid string) {
	model, ok := Cache.Get(orderUid)

	if ok == false {
		ctx.Set(r, "error", fmt.Sprintf("Заказ с OrderUid=%s не найден", orderUid))
		return
	}
	ctx.Set(r, "data", model)
}

func ModelGetAll(r *http.Request) {
	ctx.Set(r, "data", Cache.GetAll())
}

func ModelRemove(orderUid string) {
	modelDelete := `DELETE FROM models WHERE order_uid=($1)`

	_, err := Db.Exec(modelDelete, orderUid)
	if err != nil {
		log.Printf("Ошибка при удалении заказа с order_uid=%s: %s", orderUid, err)
		return
	}

	paymentDelete := `DELETE FROM payments WHERE order_uid=($1)`

	_, err = Db.Exec(paymentDelete, orderUid)
	if err != nil {
		log.Printf("Ошибка при удалении оплаты с order_uid=%s: %s", orderUid, err)
		return
	}

	deliveryDelete := `DELETE FROM deliveries WHERE order_uid=($1)`

	_, err = Db.Exec(deliveryDelete, orderUid)
	if err != nil {
		log.Printf("Ошибка при удалении доставки с order_uid=%s: %s", orderUid, err)
		return
	}

	itemsDelete := `DELETE FROM items WHERE order_uid=($1)`

	_, err = Db.Exec(itemsDelete, orderUid)
	if err != nil {
		log.Printf("Ошибка при удалении товаров с order_uid=%s: %s", orderUid, err)
	}

	Cache.Remove(orderUid)
}

func NatsInit(msg time.Time) {
	row := Db.QueryRow("SELECT COUNT(*) FROM nats WHERE host = 'localhost' ")

	var i int

	row.Scan(&i)

	if i > 0 {
		return
	}

	time := msg.Format(time.RFC822Z)

	_, err := Db.Exec(`INSERT INTO nats VALUES ($1, $2)`, "localhost", time)

	if err != nil {
		log.Fatalf("%s: ошибка при создании бэкапа NATS: %s", time, err)
	}
}

func NatsWriter(msg time.Time) {
	time := msg.Format(time.RFC822Z)

	_, err := Db.Exec(`UPDATE nats SET last = ($1) WHERE host='localhost'`, time)

	if err != nil {
		log.Fatalf("%s: ошибка при создании бэкапа NATS: %s", time, err)
	}
}

func NatsReader() time.Time {
	loaded := ""

	row := Db.QueryRow(`SELECT last FROM nats`)

	err := row.Scan(&loaded)

	if err != nil {
		log.Fatalf("Невозможно закгрузить бэкап NATS, пропущенные сообщения будут проигнорированны")
	}

	restoredTime, err := time.Parse(time.RFC822Z, loaded)

	if err != nil {
		log.Fatalf("Невозможно закгрузить бэкап NATS, пропущенные сообщения будут проигнорированны")
	}

	return restoredTime
}
