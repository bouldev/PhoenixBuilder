package script_engine

import "net/http"

type mutableHandler struct {
	HTTPHandler func(http.ResponseWriter,*http.Request)
}

func (h *mutableHandler) ServeHTTP(responseWriter http.ResponseWriter, req *http.Request) {
	h.HTTPHandler(responseWriter, req)
}
