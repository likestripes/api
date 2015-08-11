package api

import (
	"github.com/likestripes/pacific"
	"github.com/likestripes/things"
	"strconv"
)

func (state *State) Scope() (scope things.Scope) {

	scope.Context = state.Context
	scope.PersonId = state.PersonId
	scope.PersonIdStr = state.PersonIdStr
	if state.Origin.URL == "" {
		scope.Ancestors = things.ScopeToPerson(state.PersonId)
	} else {
		scope.Ancestors = ScopeToOrigin(state.Origin.URL, state.PersonId)
		scope.OriginId = state.Origin.URL
	}
	state.CurrentScope = scope
	return scope
}

func (state *State) ScopeToPerson() (scope things.Scope) {
	scope.Context = state.Context
	scope.Ancestors = things.ScopeToPerson(state.PersonId)
	return scope
}

func ScopeId(scope things.Scope) (id string) { //TODO: is this useful?

	for _, anc := range scope.Ancestors {
		if id != "" {
			id = id + "/"
		}
		if anc.KeyString != "" {
			id = id + anc.KeyString
		} else {
			id = id + strconv.FormatInt(anc.KeyInt, 10)
		}
	}

	return id
}

func ScopeToOrigin(origin_url string, person_id int64) (ancestors []pacific.Ancestor) {
	ancestors = things.ScopeToPerson(person_id)

	ancestor := pacific.Ancestor{
		Kind:      "Origin",
		KeyString: origin_url,
	}
	return append(ancestors, ancestor)
}
