package handler

import (
	"net/http"

	"github.com/zhirsch/destinykioskstatus/db"
)

type Handler interface {
	ServeHTTP(*db.User, http.ResponseWriter, *http.Request)
}
