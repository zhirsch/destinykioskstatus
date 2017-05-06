package handler

import (
	"net/http"

	"github.com/zhirsch/destinykioskstatus/src/db"
)

type Handler interface {
	ServeHTTP(*db.BungieUser, http.ResponseWriter, *http.Request)
}
