package cache

import (
	mdl "github.com/IKostarev/go-wb-l0/internal/model"
	"github.com/jmoiron/sqlx"
	"log"
	"time"
)

var Db *sqlx.DB

type Cache struct {
	models   map[string]mdl.Model
	natsLast time.Time
}

func CacheNew() *Cache {
	var c Cache

	c.models = make(map[string]mdl.Model)

	return &c
}

func (c *Cache) Append(model mdl.Model) {
	c.models[model.OrderUid] = model
}

func (c *Cache) Load() {
	modelQuery := `SELECT * FROM orders`

	rows, err := Db.Queryx(modelQuery)

	if err != nil {
		log.Printf("При поиске заказов возникла ошибка:%s", err)
		return
	}

	paymentQuery := `SELECT * FROM payments WHERE order_uid=($1)`

	deliveryQuery := `SELECT * FROM deliveries WHERE order_uid=($1)`

	itemQuery := `SELECT * FROM items WHERE order_uid=($1)`

	models := make(map[string]mdl.Model)

	for rows.Next() {
		var model mdl.Model

		err := rows.StructScan(&model)

		if err != nil {
			log.Printf("При загрузке в кэш заказа с uid=%s возникла ошибка:%s", model.OrderUid, err)
			continue
		}

		row := Db.QueryRowx(paymentQuery, model.OrderUid)

		err = row.StructScan(&model.Payment)

		if err != nil {
			log.Printf("При загрузке в кэш оплаты заказа с uid=%s возникла ошибка:%s", model.OrderUid, err)
			continue
		}

		row = Db.QueryRowx(deliveryQuery, model.OrderUid)

		err = row.StructScan(&model.Delivery)

		if err != nil {
			log.Printf("При загрузке в кэш заказа с uid=%s возникла ошибка:%s", model.OrderUid, err)
			continue
		}

		rows, err := Db.Queryx(itemQuery, model.OrderUid)

		if err != nil {
			log.Printf("При загрузке в кэш заказа с uid=%s возникла ошибка:%s", model.OrderUid, err)
			continue
		}

		var item mdl.Item

		for rows.Next() {
			err := rows.StructScan(&item)

			if err != nil {
				log.Printf("При сзагрузке в кэш товара заказа с uid=%s возникла ошибка:%s", model.OrderUid, err)
				continue
			}

			model.Items = append(model.Items, item)
		}
		models[model.OrderUid] = model
	}
	c.models = models
}

func (c Cache) GetAll() []string {
	var models []string

	for o, _ := range c.models {
		models = append(models, o)
	}

	return models
}

func (c Cache) Get(orderuid string) (model mdl.Model, ok bool) {
	model, ok = c.models[orderuid]

	if ok == true {
		model.PrepareOut()
	}

	return model, ok
}

func (c *Cache) Remove(orderuid string) {
	delete(c.models, orderuid)
}
