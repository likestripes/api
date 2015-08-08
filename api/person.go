package api

import (
	"fmt"
	"github.com/likestripes/kolkata"
	"github.com/likestripes/things"
	"net/http"
)

type Person struct {
	*kolkata.Person `datastore:"-" sql:"-" json:"-"`
	Scope           *things.Scope `datastore:"-" sql:"-" json:"-"`
	Username        string
}

func init() {
	http.HandleFunc("/bootstrap", BootstrapHandler)
	http.HandleFunc("/user/hello", Hello)
	http.HandleFunc("/user/error", Errr)
}

func Hello(w http.ResponseWriter, r *http.Request) {

	username := r.FormValue("username")
	param := r.FormValue("param")
	src_person, scope, state := ScopeAndState(w, r)

	if username != "" {
		state.InitPerson(&src_person, &scope, username)
	}

	if param != "" {
		scope := state.ScopeToPerson()
		origin := scope.Thing(param)
		state.Context.Infof(origin.ThingId)
		if redirect_to, ok := origin.Map["redirect_url"]; ok {
			http.Redirect(state.Writer, state.Request, redirect_to.(string), 301)
		}
	}
}

func Errr(w http.ResponseWriter, r *http.Request) {
	param := r.FormValue("param")
	fmt.Fprint(w, "errr... "+param)
}

func (state *State) InitPerson(src_person *kolkata.Person, scope *things.Scope, username_ar ...string) (person Person) {

	username := ""

	if len(username_ar) > 0 {
		username = username_ar[0]
	} else {

		person_thing := scope.Thing(src_person.PersonIdStr)

		if username_int, ok := person_thing.Map["Username"]; ok { //TODO: crufty
			username = username_int.(string)
			services := scope.Thing(src_person.PersonIdStr + "/services")
			if services.IsNew {
				services.Save()
			}
			state.Services = services.Map
		}
	}

	person = state.NewPerson(username, scope, src_person)

	if len(username_ar) > 0 {
		person.Save()
	}

	return person

}

func (state *State) NewPerson(username string, scope *things.Scope, src_person *kolkata.Person) (person Person) {

	person.Scope = scope
	person.Person = src_person
	person.Username = username

	return
}

func (person Person) Save() {
	thing := person.ToThing()
	thing.Save()
}

func (person Person) ToThing() (thing things.Thing) {
	thing = things.Thing{
		ThingId:  person.PersonIdStr,
		PersonId: person.PersonId,
	}
	scope := *person.Scope
	thing.Map = make(map[string]interface{})
	thing.Map["Username"] = person.Username
	thing.Scope = scope
	thing.Tags = make(map[string]things.Tag)
	thing.Tags[person.Username] = things.Tag{
		TagId: person.Username,
	}

	return
}
