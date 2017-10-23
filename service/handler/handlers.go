package handler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/vjsamuel/uploadly/service/storage"
	"github.com/vjsamuel/uploadly/service/storage/object"
	"github.com/vjsamuel/uploadly/service/cache"
	"github.com/vjsamuel/uploadly/service/common"
	"github.com/vjsamuel/uploadly/service/auth"
	"github.com/vjsamuel/uploadly/service/storage/entity"
	"github.com/vjsamuel/uploadly/service/pubsub"
	"github.com/vjsamuel/uploadly/service/memcache"
)

type handler struct {
	object storage.Storage
	entity storage.Storage
	psub   *pubsub.PubSub
	users *cache.EvictableMap
	mcache *memcache.Memcache
}

func NewHandler(users *cache.EvictableMap) *handler {
	projectId := os.Getenv("PROJECT_ID")
	bucket := os.Getenv("BUCKET")
	ctx := context.Background()

	o := object.NewObjectStorage(bucket, projectId, ctx)
	if o == nil {
		log.Fatal("Unable to create cloud storage client")
	}
	e := entity.NewEntityStorage(projectId, ctx)
	if e == nil {
		log.Fatal("Unable to create datastore client")
	}

	p := pubsub.NewPubSub(projectId, bucket, ctx)
	if p == nil {
		log.Fatal("Unable to create pubsub client")
	}

	host := os.Getenv("MEMCACHE_SERVICE_HOST")
	if host == "" {
		host = "localhost"
	}

	port := os.Getenv("MEMCACHE_SERVICE_PORT")
	if port == "" {
		port = "11211"
	}

	mcache := memcache.NewMemcacheStorage(host, port)

	return &handler{object: o, users: users, entity: e, psub: p, mcache: mcache}
}

func (h *handler) GetFiles(w http.ResponseWriter, r *http.Request) {
	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	holder := common.Holder{
		User: *usr,
	}

	if rawResp, _ := h.mcache.GetList(holder); rawResp != nil {
		bytes, _ := rawResp.([]byte)
		fmt.Fprintf(w, "%s", string(bytes))
		return
	}

	rawResp, err := h.entity.List(holder)
	if err != nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	if rawResp == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, _ := rawResp.([]common.Response)
	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Unable to get file info", http.StatusInternalServerError)
		return
	}

	holder.Object = bytes
	h.mcache.InsertList(holder)

	fmt.Fprintf(w, "%s", string(bytes))
}

func (h *handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	lenStr := r.Header.Get("Content-Length")
	var err error
	var length int64
	if lenStr != "" {
		length, err = strconv.ParseInt(lenStr, 10, 64)
		if err != nil {
			http.Error(w, "Unable to process file", http.StatusInternalServerError)
			return
		}

		if length > 1024 * 1024 * 10 {
			http.Error(w, "File size exceeded 10 MB", http.StatusBadRequest)
		}

	}
	a, b, err := r.FormFile("file")
	description := r.FormValue("description")

	if err != nil {
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		log.Printf("Unable to upload file due to error: %v\n", err)
		return
	}

	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	reader := r.Body
	bytes, _ := ioutil.ReadAll(reader)
	fmt.Println(string(bytes))


	contentType := b.Header.Get("Content-Type")
	holder := common.Holder{
		File: b.Filename,
		User: *usr,
		Object: a,
		ContentType: contentType,
		Size: length,
		Description: description,
	}

	err = h.psub.Publish(holder)
	if err != nil {
		http.Error(w, "Unable to process file", http.StatusInternalServerError)
		return
	}
	a.Close()

	err = h.entity.Insert(holder)
	if err != nil {
		http.Error(w, "Unable to process file", http.StatusInternalServerError)
		return
	}

	h.mcache.Delete(holder)
	h.mcache.DeleteList(holder)

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "%s uploaded", b.Filename)
}

func (h *handler) UpdateFile(w http.ResponseWriter, r *http.Request) {
	lenStr := r.Header.Get("Content-Length")
	var err error
	var length int64
	if lenStr != "" {
		length, err = strconv.ParseInt(lenStr, 10, 64)
		if err != nil {
			http.Error(w, "Unable to process file", http.StatusInternalServerError)
			return
		}

		if length > 1024 * 1024 * 10 {
			http.Error(w, "File size exceeded 10 MB", http.StatusBadRequest)
		}

	}
	a, b, err := r.FormFile("file")
	description := r.FormValue("description")

	if err != nil {
		http.Error(w, "Unable to upload file", http.StatusInternalServerError)
		log.Printf("Unable to upload file due to error: %v\n", err)
		return
	}

	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	contentType := b.Header.Get("Content-Type")
	holder := common.Holder{
		File: b.Filename,
		User: *usr,
		Object: a,
		ContentType: contentType,
		Size: length,
		Description: description,
	}
	if  h.entity.Exists(holder) == false {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = h.psub.Publish(holder)
	if err != nil {
		http.Error(w, "Unable to process file", http.StatusInternalServerError)
		return
	}
	a.Close()

	err = h.entity.Update(holder)
	if err != nil {
		http.Error(w, "Unable to process file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "%s uploaded", b.Filename)

	h.mcache.Delete(holder)
	h.mcache.DeleteList(holder)
}

func (h *handler) GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	holder := common.Holder{
		File: name,
		User: *usr,
	}

	exists := h.object.Exists(holder)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	rawReader, err:= h.object.Get(holder)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Unable to get file. Please try again")
		return
	}

	if rawReader == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reader, ok := rawReader.(io.ReadCloser)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Unable to get file. Please try again")
		return
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Unable to get file. Please try again")
		return
	}
	w.Header().Add("Content-Length", fmt.Sprintf("%d", len(bytes)))
	w.Header().Add("Cache-Control", "s-maxage=3600, public")
	w.Write(bytes)
	reader.Close()
}

func (h *handler) GetFileInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	holder := common.Holder{
		File: name,
		User: *usr,
	}

	if rawResp, _ := h.mcache.Get(holder); rawResp != nil {
		bytes, _ := rawResp.([]byte)
		fmt.Fprintf(w, "%s", string(bytes))
		return
	}

	rawResp, err := h.entity.Get(holder)

	if rawResp == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	resp, _ := rawResp.(common.Response)
	bytes, err := json.Marshal(resp)

	if err != nil {
		http.Error(w, "Unable to get file info", http.StatusInternalServerError)
		return
	}

	holder.Object = bytes
	h.mcache.Insert(holder)

	fmt.Fprintf(w, "%s", string(bytes))
}

func (h *handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	usr := h.getUserFromRequest(r)
	if usr == nil {
		http.Error(w, "Unable to process request", http.StatusInternalServerError)
		return
	}

	holder := common.Holder{
		File: name,
		User: *usr,
	}

	exists := h.entity.Exists(holder)
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := h.object.Delete(holder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Unable to delete file. Please try again")
		return
	}

	err = h.entity.Delete(holder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Unable to delete file metadata. Please try again")
		return
	}

	h.mcache.Delete(holder)
	h.mcache.DeleteList(holder)
	w.WriteHeader(http.StatusOK)
}

func (h *handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ok"))
}

func (h *handler) getUserFromRequest(r *http.Request) *common.User{
	token := auth.GetAuthToken(r)
	return h.users.Get(token)
}


