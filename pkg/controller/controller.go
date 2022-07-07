package controller

import (
	"github.com/gorilla/mux"
	"gitlab.com/IKostarev/go-wb-l0/pkg/service"
	"net/http"
)

func ModelGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	service.ModelGet(r, vars["orderuid"])
}

func ModelGetAll(w http.ResponseWriter, r *http.Request) {
	service.ModelGetAll(r)
}
