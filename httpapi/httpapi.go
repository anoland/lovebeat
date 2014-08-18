package httpapi

import (
	"encoding/json"
	"github.com/boivie/lovebeat-go/service"
	"github.com/gorilla/mux"
	"github.com/op/go-logging"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	svcs   *service.Services
	client service.ServiceIf
)

func now() int64 { return time.Now().Unix() }

var log = logging.MustGetLogger("lovebeat")

func ServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	var err = r.ParseForm()
	if err != nil {
		log.Error("error parsing form ", err)
		return
	}

	var errtmo, warntmo = r.FormValue("err-tmo"), r.FormValue("warn-tmo")

	client.Beat(name)

	if val, err := strconv.Atoi(errtmo); err == nil {
		client.SetErrorTimeout(name, val)
	}

	if val, err := strconv.Atoi(warntmo); err == nil {
		client.SetWarningTimeout(name, val)
	}

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteService(name)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func CreateViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	view_name := params["name"]
	var expr = r.FormValue("regexp")
	if expr == "" {
		log.Error("No regexp provided")
		return
	}

	client.CreateOrUpdateView(view_name, expr)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteViewHandler(c http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	name := params["name"]

	client.DeleteView(name)

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

type JsonView struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

func GetViewsHandler(c http.ResponseWriter, r *http.Request) {
	var ret = make([]JsonView, 0)
	for _, v := range client.GetViews() {
		js := JsonView{
			Name:  v.Name,
			State: v.State,
		}
		ret = append(ret, js)
	}
	var encoded, _ = json.MarshalIndent(ret, "", "  ")

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

type JsonService struct {
	Name           string `json:"name"`
	LastBeat       int64  `json:"last_beat"`
	WarningTimeout int64  `json:"warning_timeout"`
	ErrorTimeout   int64  `json:"error_timeout"`
	State          string `json:"state"`
}

func GetServicesHandler(c http.ResponseWriter, r *http.Request) {
	viewName := "all"

	if val, ok := r.URL.Query()["view"]; ok {
		viewName = val[0]
	}

	var ret = make([]JsonService, 0)
	for _, s := range client.GetServices(viewName) {
		js := JsonService{
			Name:           s.Name,
			LastBeat:       s.LastBeat,
			WarningTimeout: s.WarningTimeout,
			ErrorTimeout:   s.ErrorTimeout,
			State:          s.State,
		}
		ret = append(ret, js)
	}
	var encoded, _ = json.MarshalIndent(ret, "", "  ")

	c.Header().Add("Content-Type", "text/plain")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func Register(rtr *mux.Router, client_ service.ServiceIf) {
	client = client_
	rtr.HandleFunc("/api/services/", GetServicesHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:[a-z0-9.]+}", ServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:[a-z0-9.]+}", DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/api/views/", GetViewsHandler).Methods("GET")
	rtr.HandleFunc("/api/views/{name:[a-z0-9.]+}", CreateViewHandler).Methods("POST")
	rtr.HandleFunc("/api/views/{name:[a-z0-9.]+}", DeleteViewHandler).Methods("DELETE")
}