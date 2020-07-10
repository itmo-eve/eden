package server

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/lf-edge/eden/eserver/api"
	"github.com/lf-edge/eden/eserver/pkg/manager"
	"net/http"
)

type adminHandler struct {
	manager *manager.EServerManager
}

func (h *adminHandler) list(w http.ResponseWriter, r *http.Request) {
	files := h.manager.ListFileNames()
	w.Header().Add(contentType, mimeTextPlain)
	w.WriteHeader(http.StatusOK)
	for _, value := range files {
		fileName := bytes.NewBufferString(value + "\n")
		if _, err := fileName.WriteTo(w); err != nil {
			wrapError(err, w)
			return
		}
	}
}

func (h *adminHandler) addFromUrl(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data api.UrlArg
	err := decoder.Decode(&data)
	if err != nil {
		wrapError(err, w)
		return
	}
	name, err := h.manager.AddFile(data.Url)
	if err != nil {
		wrapError(err, w)
		return
	}
	w.Header().Add(contentType, mimeTextPlain)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(name))
}

func (h *adminHandler) getFileStatus(w http.ResponseWriter, r *http.Request) {
	u := mux.Vars(r)["filename"]
	fileInfo, err := h.manager.GetFileInfo(u)
	if err != nil {
		wrapError(err, w)
		return
	}
	out, err := json.Marshal(fileInfo)
	if err != nil {
		wrapError(err, w)
		return
	}
	w.Header().Add(contentType, mimeTextPlain)
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
