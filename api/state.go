package api

import (
	"github.com/likestripes/kolkata"
	"github.com/likestripes/pacific"
	"github.com/likestripes/things"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var public_host = "lsapi-vm.appspot.com"
var public_url  = "https://"+public_host

type State struct {
	Bootstrap    string
	Writer       http.ResponseWriter
	Request      *http.Request
	Context      pacific.Context
	CurrentScope things.Scope
	AcceptType   string
	ContentType  string
	Path         string
	PersonId     int64
	PersonIdStr  string
	Host         string
	Services     map[string]interface{}
	Person       kolkata.Person
	ThePerson    Person
	NewUser      bool
	Origin       Origin
	OriginStr    string
	Dev          bool
}

func CreateState(w http.ResponseWriter, r *http.Request) (state State) {

	w.Header().Set("Content-Language", "en")

	c := pacific.NewContext(r)

	rand.Seed(int64(time.Now().Nanosecond()))

	accept := r.Header.Get("Accept")
	content := r.Header.Get("Content-Type")

	origin := r.Referer()

	if origin == "" {
		if val, ok := r.Header["Origin"]; ok {
			origin = val[0]
		} else {
			origin = public_url
		}
	}

	origin = HostAndPort(origin)

	state = State{
		AcceptType:  accept,
		ContentType: content,
		Writer:      w,
		Request:     r,
		Context:     c,
		OriginStr:   origin,
		Host:        public_url,
		Path:        strings.ToLower(r.URL.Path),
	}

	if pacific.IsDevAppServer() {
		state.Host = "http://localhost:8080"
		state.Dev = true
		if state.OriginStr == public_host {
			state.OriginStr = "localhost:8080"
		}
	}

	state.Origin = state.SetOrigin()

	return
}
