package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/boivie/lovebeat/algorithms"
	"github.com/boivie/lovebeat/model"
	"github.com/boivie/lovebeat/service"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"strconv"
)

var version string

func parseTimeout(tmo string) int64 {
	if tmo == "auto" {
		return model.TIMEOUT_AUTO
	} else {
		val, _ := strconv.Atoi(tmo)
		return int64(val) * 1000
	}
}

func replyJson(c http.ResponseWriter, js interface{}) {
	var encoded, _ = json.MarshalIndent(js, "", "  ")
	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
	c.Write(encoded)
	io.WriteString(c, "\n")
}

func ServiceHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]

	update := &model.Update{Ts: now(), Service: name}

	// Timeout as query parameter
	if val, ok := r.URL.Query()["timeout"]; ok {
		update.SetTimeout = &model.SetTimeout{Timeout: parseTimeout(val[0])}
	}

	if r.Header.Get("Content-Type") == "application/json" {
		decoder := json.NewDecoder(r.Body)
		var t struct {
			Timeout *int64 `json:"timeout"`
			Beat    *bool  `json:"beat"`
		}
		err := decoder.Decode(&t)
		if err == nil && t.Timeout != nil {
			if *t.Timeout > 0 {
				update.SetTimeout = &model.SetTimeout{Timeout: *t.Timeout * 1000}
			} else {
				update.SetTimeout = &model.SetTimeout{Timeout: *t.Timeout}
			}
		}
		if err == nil && (t.Beat == nil || *t.Beat == true) {
			update.Beat = &model.Beat{}
		}
	} else {
		var err = r.ParseForm()
		if err != nil {
			log.Errorf("error parsing form: %v", err)
			return
		}

		if r.FormValue("timeout") != "" {
			update.SetTimeout = &model.SetTimeout{Timeout: parseTimeout(r.FormValue("timeout"))}
		}

		update.Beat = &model.Beat{}
	}

	client.Update(update)

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func MuteServiceHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&model.Update{Ts: now(), Service: name, MuteService: &model.MuteService{Muted: true}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func UnmuteServiceHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&model.Update{Ts: now(), Service: name, MuteService: &model.MuteService{Muted: false}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func DeleteServiceHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]

	client.Update(&model.Update{Ts: now(), Service: name, DeleteService: &model.DeleteService{}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

func GetAlarmsHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	alarms := client.GetAlarms()
	replyJson(c, struct {
		Alarms []model.Alarm `json:"alarms"`
		Now    int64         `json:"now"`
	}{alarms, now()})
}

func GetAlarmHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]
	v := client.GetAlarm(name)
	if v == nil {
		http.NotFound(c, r)
		return
	}
	services := client.GetServicesInAlarm(name)
	retServices := make([]HttpApiService, len(services))
	for i, s := range services {
		retServices[i] = ToHttpService(s)
	}

	replyJson(c, struct {
		Alarm    *model.Alarm     `json:"alarm"`
		Services []HttpApiService `json:"services"`
		Now      int64            `json:"now"`
	}{v, retServices, now()})
}

func DeleteAlarmHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)

	client.Update(&model.Update{Ts: now(), Alarm: params["name"], DeleteAlarm: &model.DeleteAlarm{}})

	c.Header().Add("Content-Type", "application/json")
	c.Header().Add("Content-Length", "3")
	io.WriteString(c, "{}\n")
}

type JsonAlarmRef struct {
	Name string `json:"name"`
}

type HttpApiService struct {
	model.Service
	Analysis *algorithms.BeatAnalysis `json:"analysis,omitempty"`
}

func ToHttpService(s model.Service) HttpApiService {
	analysis := algorithms.AnalyzeBeats(s.BeatHistory)
	s.BeatHistory = nil
	return HttpApiService{s, analysis}
}

func GetServicesHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)

	var now = now()
	services := client.GetServices()
	retServices := make([]HttpApiService, len(services))
	for i, s := range services {
		retServices[i] = ToHttpService(s)
	}
	replyJson(c, struct {
		Services []HttpApiService `json:"services"`
		Now      int64            `json:"now"`
	}{retServices, now})
}

func GetServiceHandler(c http.ResponseWriter, r *http.Request) {
	log.Debugf("%s %s", r.Method, r.RequestURI)
	params := mux.Vars(r)
	name := params["name"]

	var now = now()
	var s = client.GetService(name)
	if s == nil {
		http.NotFound(c, r)
		return
	}

	replyJson(c, struct {
		Service HttpApiService `json:"service"`
		Now     int64          `json:"now"`
	}{ToHttpService(*s), now})
}

func StatusHandler(c http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.RequestURI)
	alarmName := ""

	params := mux.Vars(req)
	if val, ok := params["name"]; ok {
		alarmName = val
	}

	var buffer bytes.Buffer
	var services = client.GetServicesInAlarm(alarmName)
	var errors, ok = 0, 0
	for _, s := range services {
		if s.State == model.StateError {
			errors++
		} else {
			ok++
		}
	}
	if req.Header.Get("Accept") == "application/json" {
		ret := struct {
			NumOk    int  `json:"num_ok"`
			NumError int  `json:"num_error"`
			HasError bool `json:"has_error"`
			Good     bool `json:"good"`
		}{
			ok, errors,
			errors > 0, errors == 0,
		}
		var encoded, _ = json.MarshalIndent(ret, "", "  ")

		c.Header().Add("Content-Type", "application/json")
		c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
		c.Write(encoded)
		io.WriteString(c, "\n")
	} else {
		buffer.WriteString(fmt.Sprintf("num_ok %d\nnum_error %d\n", ok, errors))
		buffer.WriteString(fmt.Sprintf("has_error %t\ngood %t\n", errors > 0, errors == 0))
		body := buffer.String()
		c.Header().Add("Content-Type", "text/plain")
		c.Header().Add("Content-Length", strconv.Itoa(len(body)))
		io.WriteString(c, body)
	}
}

func VersionHandler(c http.ResponseWriter, req *http.Request) {
	log.Debugf("%s %s", req.Method, req.RequestURI)
	if req.Header.Get("Accept") == "application/json" {
		ret := struct {
			Version string `json:"version"`
		}{
			version,
		}
		var encoded, _ = json.MarshalIndent(ret, "", "  ")

		c.Header().Add("Content-Type", "application/json")
		c.Header().Add("Content-Length", strconv.Itoa(len(encoded)+1))
		c.Write(encoded)
		io.WriteString(c, "\n")
	} else {
		c.Header().Add("Content-Type", "text/plain")
		c.Header().Add("Content-Length", strconv.Itoa(len(version)+1))
		io.WriteString(c, version+"\n")
	}
}

func AddEndpoints(rtr *mux.Router, version_ string) {
	version = version_
	rtr.HandleFunc("/api/services", GetServicesHandler).Methods("GET")
	rtr.HandleFunc("/api/services/", GetServicesHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", ServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}/mute", MuteServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}/unmute", UnmuteServiceHandler).Methods("POST")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", GetServiceHandler).Methods("GET")
	rtr.HandleFunc("/api/services/{name:"+service.ServiceNamePattern+"}", DeleteServiceHandler).Methods("DELETE")
	rtr.HandleFunc("/api/alarms", GetAlarmsHandler).Methods("GET")
	rtr.HandleFunc("/api/alarms/", GetAlarmsHandler).Methods("GET")
	rtr.HandleFunc("/api/alarms/{name:"+service.ServiceNamePattern+"}", GetAlarmHandler).Methods("GET")
	rtr.HandleFunc("/api/alarms/{name:"+service.ServiceNamePattern+"}", DeleteAlarmHandler).Methods("DELETE")
	rtr.HandleFunc("/api/alarms/{name:"+service.ServiceNamePattern+"}/status", StatusHandler).Methods("GET")
	rtr.HandleFunc("/api/status", StatusHandler).Methods("GET")
	rtr.HandleFunc("/version", VersionHandler).Methods("GET")
}
