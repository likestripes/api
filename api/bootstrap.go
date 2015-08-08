package api

import (
	"github.com/likestripes/kolkata"
	"net/http"
)

func BootstrapHandler(w http.ResponseWriter, r *http.Request) {
	state := CreateState(w, r)

	src_person, _ := kolkata.Current(w, r)
	state.Person = src_person
	state.PersonId = src_person.PersonId
	state.PersonIdStr = src_person.PersonIdStr

	state.Context.Infof("BOOTSTRAPPING with userid: " + src_person.PersonIdStr)

	origin := state.NewOrigin("lsapi-vm.appspot.com") //self appengine-vm
	origin.Save()
	origin = state.NewOrigin("api.likestripes.com") //self appengine
	origin.Save()
	origin = state.NewOrigin("localhost:8080") //self local
	origin.Save()
	origin = state.NewOrigin("apps.likestripes.com") //apps appengine
	origin.Save()
	origin = state.NewOrigin("localhost:8004") //pages local
	origin.Save()
	origin = state.NewOrigin("localhost:8001") //apps local
	origin.Save()
	origin = state.NewOrigin("localhost:8005") //today local
	origin.Save()
}
