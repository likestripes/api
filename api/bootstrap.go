package api

import (
	"github.com/likestripes/kolkata"
	"github.com/likestripes/pacific"
	"github.com/likestripes/things"
	"net/http"
)

func BootstrapHandler(w http.ResponseWriter, r *http.Request) {
	state := CreateState(w, r)

	pacific.AutoMigrate(state.Context, "Person", "person_id", &kolkata.Person{})
	pacific.AutoMigrate(state.Context, "Session", "session_id", &kolkata.Session{})
	pacific.AutoMigrate(state.Context, "SignIn", "signin_id", &kolkata.SignIn{})
	pacific.AutoMigrate(state.Context, "Thing", "thing_id", &things.Thing{})
	pacific.AutoMigrate(state.Context, "SharedThing", "object_id", &things.Share{})
	pacific.AutoMigrate(state.Context, "SharedTag", "object_id", &things.Share{})

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
