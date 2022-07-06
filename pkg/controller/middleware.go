package controller

import (
	"encoding/json"
	ctx "github.com/gorilla/context"
	"log"
	"net/http"
)

func MiddlewarePackData(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)

		data := ctx.Get(r, "data")
		ctx.Delete(r, "data")

		if ctx.Get(r, "error") != nil {
			w.WriteHeader(404)
			w.Write([]byte(ctx.Get(r, "error").(string)))
			ctx.Delete(r, "error")
			return
		}

		if data == nil {
			w.WriteHeader(404)
			w.Write([]byte("Информация не найдена"))
			return
		}

		responseBody, _ := json.MarshalIndent(data, " ", "\t")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(responseBody)

		if err != nil {
			log.Println("Ошибка при отправке ответа:", err)
		}
	})
}