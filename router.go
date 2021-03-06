package main

import (
	"encoding/json"
	"github.com/awsmsrc/llog"
	"github.com/awsmsrc/params"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strconv"
)

func InitRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/accounts/{account_id:[0-9]+}/test", TestHandler)
	r.HandleFunc("/accounts/{account_id:[0-9]+}/registrations", RegistrationsHandler)
	r.HandleFunc("/accounts/{account_id:[0-9]+}/events", EventsHandler)
	r.HandleFunc("/accounts/{account_id:[0-9]+}/attempts", AttemptsHandler)
	r.NotFoundHandler = http.HandlerFunc(Render404)
	return r
}

func TestHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "POST":
		p, err := params.NewFromJsonRequestBody(req)
		if err != nil {
			Render400(w)
			return
		}
		p.RequireString("data")
		p.RequireString("url")
		err = p.Validate()
		if err != nil {
			RenderBadParam400(w, err.(*params.Error).Params[0])
			return
		}
		m := p.Map()
		t := &Tester{
			Url: m["url"].(string),
			Data: m["data"].(string),
		}
		attempt, err := t.Execute()
		if err != nil {
			Render(
				w,
				map[string]Error{
					"error":Error{
						Message: "URL unavailable",
					},
				},
				200,
			)
			return
		}
		Render(w, attempt, 200)
	default:
		Render405(w)
	}
}

//registration crud
func RegistrationsHandler(w http.ResponseWriter, req *http.Request) {
	account_id, _ := strconv.Atoi(mux.Vars(req)["account_id"])
	switch req.Method {
	case "GET":
		registrations := GetRegistrations(account_id)
		Render(w, registrations, 200)
	case "POST":
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			Render400(w)
			return
		}
		registration := new(Registration)
		err = json.Unmarshal(body, registration)
		if err != nil {
			Render400(w)
			return
		}
		llog.Successf("struct created: %+v", registration)
		_, err = db.Create(registration)
		if err != nil {
			llog.Error(err)
		}
		w.WriteHeader(201)
	case "DELETE":
		llog.FATAL("TODO")
	default:
		Render405(w)
	}
}

//ACCOUNT EVENT CRUD
func EventsHandler(w http.ResponseWriter, req *http.Request) {
	account_id, _ := strconv.Atoi(mux.Vars(req)["account_id"])
	switch req.Method {
	case "GET":
		Render(w, GetEvents(account_id), 200)
	case "POST":
		llog.Infof("Params are :%+v", req)
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			Render400(w)
			return
		}
		llog.Debug(string(body))
		event := new(Event)
		err = json.Unmarshal(body, event)
		if err != nil {
			Render400(w)
			return
		}
		event.AccountId = account_id

		*event, err = RegisterEvent(*event)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		Render(w, event, 201)
	default:
		Render405(w)
	}
}

// CRUD ATTEMPTS
//#TODO this shows all attempts
func AttemptsHandler(w http.ResponseWriter, req *http.Request) {
	account_id, _ := strconv.Atoi(mux.Vars(req)["account_id"])
	switch req.Method {
	case "GET":
		Render(w, GetAttempts(account_id), 200)
	default:
		Render405(w)
	}
}
