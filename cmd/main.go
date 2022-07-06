package main

import (
	"encoding/json"
	"github.com/IKostarev/go-wb-l0/cache"
	"github.com/IKostarev/go-wb-l0/internal/model"
	"github.com/IKostarev/go-wb-l0/pkg/controller"
	"github.com/IKostarev/go-wb-l0/pkg/service"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/stan.go"
	"log"
	"net/http"
	"strings"
	"time"
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

func natsReceiver() {
	timestamp, _ := time.Parse(time.RFC822Z, "08 Jul 22 12:00 +0500")

	service.NatsInit(timestamp)

	stantime := service.NatsReader()

	log.Println("NATS запущен, дата и время последнего прочтенного сообщения:", stantime.Format(time.RFC822Z))

	sc, err := stan.Connect("test-cluster", "go-nats-streaming-json-receiver")

	if err != nil {
		log.Fatal("NATS не запущен", err)
	}

	sub, _ := sc.Subscribe("models", func(m *stan.Msg) {
		service.NatsWriter(time.Now())

		var model model.Model

		err := json.Unmarshal(m.Data, &model)

		if err != nil {
			log.Println("Невозможно прочитать сообщение из NATS")
		} else {
			model.PrepareIn()
			err := service.ModelAdd(model)

			if err != nil {
				if strings.Contains(err.Error(), "повторяющееся значение ключа") == false {
					log.Println(err)
				}
			}
		}
	}, stan.StartAtTime(stantime))

	time.Sleep(10 * time.Minute)

	sub.Unsubscribe()
}

func main() {
	Cache := cache.CacheNew()
	service.Cache = Cache
	mR := mux.NewRouter()
	mR.StrictSlash(true)

	db = service.DBInit(dbConf)
	db.SetMaxOpenConns(100)
	service.Db = db
	cache.Db = db
	service.DBInit(dbConf)

	Cache.Load()

	go natsReceiver()

	getModel := http.HandlerFunc(controller.ModelGet)
	getModelAll := http.HandlerFunc(controller.ModelGetAll)

	mR.Handle("/api/v1/models/{orderuid}", controller.MiddlewarePackData(getModel))
	mR.Handle("/api/v1/models/", controller.MiddlewarePackData(getModelAll)).Methods("GET")

	log.Fatal(http.ListenAndServe("localhost:8080", mR))
}
