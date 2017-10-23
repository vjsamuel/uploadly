package main

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/vjsamuel/uploadly/service/handler"
	"github.com/vjsamuel/uploadly/service/cache"
	"github.com/vjsamuel/uploadly/service/auth"
)

var users = cache.NewEvictableMap(100, time.Minute)

func main() {
	h := handler.NewHandler(users)
	a := auth.NewAuthHandler(users)

	r := mux.NewRouter()

	v1 := r.PathPrefix("/api/v1").Subrouter()
	v1.Path("/files").Handler(a.AuthenticatedHandler(cache.NoCacheHandler(h.GetFiles))).Methods("GET")
	v1.Path("/files").Handler(a.AuthenticatedHandler(h.UploadFile)).Methods("POST")
	v1.Path("/files").Handler(a.AuthenticatedHandler(h.UpdateFile)).Methods("PUT")

	file := v1.PathPrefix("/file").Subrouter()
	file.Path("/{name}").Handler(a.AuthenticatedHandler(h.GetFile)).Methods("GET")
	file.Path("/{name}").Handler(a.AuthenticatedHandler(h.DeleteFile)).Methods("DELETE")

	pages := v1.PathPrefix("/file/{name}").Subrouter()
	pages.Path("/info").Handler(a.AuthenticatedHandler(cache.NoCacheHandler(h.GetFileInfo))).Methods("GET")

	r.Methods("GET").Path("/_ah/health").Handler(cache.NoCacheHandler(h.HealthCheck))

	fs := http.FileServer(http.Dir("../webapp"))
	r.Methods("GET").PathPrefix("/").Handler(fs)

	http.ListenAndServe(":8080", r)
}
