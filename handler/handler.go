package handler

import (
	"net/http"

	"github.com/zhirsch/destinykioskstatus/db"
)

type Handler interface {
	ServeHTTP(*db.BungieUser, http.ResponseWriter, *http.Request)
}
