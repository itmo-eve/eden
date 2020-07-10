package server

import "net/http"

const (
	contentType   = "Content-Type"
	mimeTextPlain = "text/plain"
)

func wrapError(err error, w http.ResponseWriter) {
	w.Header().Add(contentType, mimeTextPlain)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}
