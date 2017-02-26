package handler

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type StackTraceMiddlewareHandler struct {
	Handler http.Handler
}

func (h StackTraceMiddlewareHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if p := recover(); p != nil {
			body := fmt.Sprintf("%v\n\n%v", p, string(debug.Stack()))
			http.Error(w, body, http.StatusInternalServerError)
		}
	}()
	h.Handler.ServeHTTP(w, r)
}
