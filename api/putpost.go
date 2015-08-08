package api

import (
	"encoding/json"
	"github.com/likestripes/things"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func KVfromRequestForm(form url.Values, highlights ...string) (kv map[string]interface{}, highlight interface{}) {

	kv = make(map[string]interface{})

	for k, v := range form {
		if len(highlights) > 0 {
			for _, h := range highlights {
				if h == k {
					highlight = v[0]
				}
			}
		}
		kv[k] = v[0]
	}

	return
}

func KVfromRequestJSON(request_body io.ReadCloser, highlights ...string) (kv map[string]interface{}, highlight interface{}) {

	kv = make(map[string]interface{})

	body, _ := ioutil.ReadAll(request_body)
	post_json := make(map[string]interface{})
	json.Unmarshal(body, &post_json)

	for k, v := range post_json {
		if len(highlights) > 0 {
			for _, h := range highlights {
				if h == k {
					highlight = v
				}
			}
		}
		kv[k] = v
	}

	return
}

func (state State) PutPostHandler() (thing things.Thing, err error) {

	scope := state.CurrentScope
	request := state.Request

	additions := make(map[string]interface{})

	thing_id := state.Path[1:]

	var kv_thing_id interface{}
	if state.ContentType[:16] == "application/json" {
		additions, kv_thing_id = KVfromRequestJSON(request.Body, "thing_id", "ThingId")
	} else {
		additions, kv_thing_id = KVfromRequestForm(request.Form, "thing_id", "ThingId")
	}

	if kv_thing_id != nil {
		thing_id = kv_thing_id.(string)
	}

	state.Context.Infof("PutPostHandler path: " + state.Path + " thing_id:" + thing_id)

	if thing_id == "" {
		thing_id = newColor()
	}
	scope_prefix := ScopeId(scope)
	if len(thing_id) < len(scope_prefix) || thing_id[0:len(scope_prefix)] != scope_prefix {
		thing_id = ScopeId(scope) + "/" + thing_id
	}

	if state.Path == "/origin" {
		if state.ThePerson.Username == "travis" {
			scope = things.Scope{
				Context: state.Context,
			}
			if url, ok := additions["URL"]; ok && url.(string) != "" {
				thing_id = "origin/" + url.(string)
				state.Context.Infof("origin: " + thing_id)
			} else {
				state.Context.Infof("Not saving blank URL origin.")
				return
			}
		}
	}

	thing = scope.Thing(thing_id)

	for k, v := range additions {
		thing.Map[k] = v
	}

	if ts, ok := additions["timestamp"].(float64); ok {
		thing.Updated = time.Unix(int64(ts), 0)
	}

	thing.PersonId = state.PersonId
	thing.PersonIdStr = state.PersonIdStr
	tag := scope.Tag(state.PersonIdStr)
	thing.Tags[tag.TagId] = tag

	thing.TagsFromString("/", state.Path, thing_id)
	thing.TagsFromString(",", additions["tags"], request.FormValue("tags"))

	thing.Save()

	return
}

func newColor() string {
	red := strconv.FormatInt(int64(rand.Intn(255)), 16)
	blue := strconv.FormatInt(int64(rand.Intn(255)), 16)
	green := strconv.FormatInt(int64(rand.Intn(255)), 16)
	color := string(red + blue + green)
	if len(color) < 6 {
		color = color + strings.Repeat("0", 6-len(color))
	}
	return color
}
