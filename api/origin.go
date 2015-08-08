package api

import (
	"github.com/likestripes/things"
	"net/url"
)

type Origin struct {
	Scope         *things.Scope `datastore:"-" sql:"-" json:"-"`
	OriginId      string
	PersonId      int64
	URL           string
	RedirectUrl   string
	Nil           bool
	*things.Thing `datastore:"-" sql:"-" json:"-"`
}

func (state State) SetOrigin() (origin Origin) {

	scope := things.Scope{
		Context: state.Context,
	}

	thing := scope.Thing("origin/" + state.OriginStr)

	if url, ok := thing.Map["URL"]; ok {
		origin.URL = url.(string)
		origin.OriginId = thing.ThingId
		if redirect_url, redirect_ok := thing.Map["redirect_url"]; redirect_ok {
			origin.RedirectUrl = redirect_url.(string)
		}
		origin.Thing = &thing
	} else {
		origin.Nil = true
	}

	return origin
}

func (state State) NewOrigin(host string) (origin Origin) {

	scope := things.Scope{
		Context: state.Context,
	}
	origin.Scope = &scope
	origin.PersonId = state.PersonId
	origin.URL = host
	origin.OriginId = "origin/" + host
	origin.RedirectUrl = "http://" + host
	thing := origin.toThing()
	origin.Thing = thing
	return
}

func (origin Origin) Kv(key string) (value string) {
	if val, ok := origin.Map[key]; ok {
		value = val.(string)
	}
	return value
}

func (origin Origin) Save() {
	thing := origin.toThing()
	thing.Save()
}

func (origin Origin) toThing() *things.Thing {

	var thing things.Thing

	if origin.Thing != nil {
		thing = *origin.Thing
	} else {
		thing = things.Thing{
			ThingId:  origin.OriginId,
			PersonId: origin.PersonId,
		}
		thing.Tags = make(map[string]things.Tag)
		thing.Tags["origin"] = things.Tag{
			TagId: "origin",
		}
	}
	if thing.Map == nil {
		thing.Map = make(map[string]interface{})
	}

	thing.Map["URL"] = origin.URL
	thing.Map["redirect_url"] = origin.RedirectUrl
	thing.Scope = *origin.Scope

	return &thing
}

func HostAndPort(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		panic(err)
	}
	return u.Host
}
