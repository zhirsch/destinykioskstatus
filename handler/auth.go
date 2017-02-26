package handler

import (
	"net/http"

	"github.com/zhirsch/destinykioskstatus/server"
)

type AuthenticationMiddlewareHandler struct {
	Server  *server.Server
	Handler Handler
}

func (h AuthenticationMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u, err := h.Server.GetUser(w, r)
	if err != nil {
		if err != server.ErrNeedAuth {
			panic(err)
		}
		return
	}
	h.Handler.ServeHTTP(u, w, r)
}
