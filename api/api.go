package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	_ "github.com/likestripes/calcutta"
	"github.com/likestripes/kolkata"
	"github.com/likestripes/moitessier"
	"github.com/likestripes/pacific"
	"github.com/likestripes/things"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	if pacific.SupportsWS() {
		go func() {
			http.ListenAndServe(":8443", &WS{})
		}()
	}

	http.HandleFunc("/", Router)
}

type WS struct{}

func (m *WS) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	state := CreateState(w, r)
	src_person, err := kolkata.Current(w, r)
	state.Person = src_person
	state.PersonId = src_person.PersonId
	state.PersonIdStr = src_person.PersonIdStr

	conditions := moitessier.Conditions{&state.Context, src_person, moitessier.Private, state.Path[1:]}
	listener := conditions.NewListener()
	err = listener.ListenUntil(ws_listener, ws_responder, time.Now(), 1000, 300000, ws_state{conn})
	timeout := time.Now().Add(time.Second * 5)

	if err != nil {
		fmt.Println(err.Error())
		conn.WriteControl(websocket.CloseInternalServerErr, []byte(err.Error()), timeout)
	}
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	listener.Close()
}

func ws_listener(listener *moitessier.Listener, args []interface{}) error {

	arg := args[0].(ws_state)

	err_ch := make(chan error, 1)
	ch := make(chan []byte, 1)
	timeout := make(chan bool, 1)

	go func() {
		_, p, err := arg.conn.ReadMessage()
		err_ch <- err
		ch <- p
	}()

	go func() {
		time.Sleep(2 * time.Millisecond)
		timeout <- true
	}()

	select {
	case err := <-err_ch:
		if err != nil {
			return err
		}
	case p := <-ch:
		if string(p) == ":close" {
			arg.conn.WriteMessage(websocket.TextMessage, []byte("closing listener and connection"))
			return errors.New("closing connection")
		}
	case <-timeout:
	}

	return nil

}

func ws_responder(msgs []moitessier.Message, args []interface{}) error {
	arg := args[0].(ws_state)
	for _, msg := range msgs {
		if err := arg.conn.WriteMessage(websocket.TextMessage, []byte(msg.Text)); err != nil {
			return err
		}
	}
	return nil
}

type ws_state struct {
	conn *websocket.Conn
}

func ScopeAndState(w http.ResponseWriter, r *http.Request) (src_person kolkata.Person, scope things.Scope, state_ptr State) {

	state := CreateState(w, r)

	src_person, err := kolkata.Current(w, r)

	if err != nil {
		state.Context.Errorf(err.Error())
	}

	state.Person = src_person
	state.PersonId = src_person.PersonId
	state.PersonIdStr = src_person.PersonIdStr

	scope = state.Scope()

	if state.Origin.Nil != true {
		set_headers(w, "http://"+state.OriginStr)
	}

	state.Context.Infof("requested: " + state.Host)
	state.Context.Infof("requested: " + state.OriginStr)
	state.Context.Infof("requested: " + state.Request.Method)

	return src_person, scope, state
}

func Router(w http.ResponseWriter, r *http.Request) {

	src_person, scope, state := ScopeAndState(w, r)

	if state.Origin.Nil {
		return
	}

	if r.Method == "OPTIONS" {
		state.Context.Infof("OPTIONS")
		if state.AcceptType == "application/json" {
			if tags := scope.Tags(); len(tags) > 0 {
				response_json, _ := json.Marshal(tags)
				fmt.Fprint(w, string(response_json))
			}
		}
		return
	}

	state.Context.Infof("OPTIONS")
	person := state.InitPerson(&src_person, &scope)
	state.ThePerson = person

	if state.Request.Method != "GET" {
		thing, err := state.PutPostHandler()
		if err != nil {
			state.Context.Errorf(err.Error())
		}
		response, _ := json.Marshal(thing.Map)
		fmt.Fprint(state.Writer, string(response))
		return
	}

	tags := []string{person.PersonIdStr}
	fields := []string{}
	scope_id := ScopeId(scope)

	if state.Path == "/origins" {
		state.Context.Infof("origins get req: " + person.Username)
		if person.Username == "travis" {
			scope = things.Scope{
				Context: state.Context,
			}
			tags = []string{"origin"}
		} else {
			return
		}
	} else {
		if len(state.Path) > 1 {
			for _, tag := range strings.Split(state.Path[1:], "/") {
				state.Context.Infof("tags for lookup: " + scope_id + "/" + tag)
				// tags = append(tags, scope_id+"/"+tag)
				if string(tag[0]) == ":" {
					fields = append(fields, tag[1:])
				} else {
					tags = append(tags, tag)
				}
			}
		}
	}
	things_ar := []things.Thing{}

	state.Context.Infof("normal origin-scoped lookup")
	things_ar = scope.Things(tags)

	things_json, _ := json.Marshal(BackboneJSON(things_ar, state, fields))

	if state.AcceptType == "application/json" {
		fmt.Fprint(state.Writer, string(things_json))
		return
	} else {
		fmt.Fprint(state.Writer, "Router\n")
		fmt.Fprint(state.Writer, state.Origin.URL+"\n")
		fmt.Fprint(state.Writer, "Person: "+person.PersonIdStr+"\n")
		fmt.Fprint(state.Writer, "Count: "+strconv.Itoa(len(things_ar))+"\n")
		if len(things_ar) > 0 {
			for _, thing := range things_ar {
				fmt.Fprint(state.Writer, "Thing: "+thing.ThingId+"\n")
			}
		}
		fmt.Fprint(state.Writer, string(things_json))
	}
	state.Context.Infof("END REQUEST")
}

func BackboneJSON(things []things.Thing, state State, fields []string) interface{} {

	if len(things) == 1 {
		val := make(map[string]interface{})
		if len(fields) == 0 {
			val = things[0].Map
		} else {
			for _, key := range fields {
				val[key] = things[0].Map[key]
			}
		}
		val["ThingId"] = things[0].ThingId
		return val
	} else if len(things) > 1 {
		json_map := make([]map[string]interface{}, 0)
		for _, thing := range things {
			if thing.ThingId != "" {
				add_thing := make(map[string]interface{})
				add_thing["id"] = thing.ThingId
				add_thing["ThingId"] = thing.ThingId
				if len(fields) == 0 {
					for k, v := range thing.Map {
						k = strings.Replace(k, ".", "", -1)
						add_thing[k] = v
					}
				} else {
					for _, key := range fields {
						add_thing[key] = thing.Map[key]
					}
				}
				json_map = append(json_map, add_thing)
			}
		}
		return json_map
	}

	state.Context.Infof("NO THINGS")
	return nil
}

func set_headers(w http.ResponseWriter, allow_origin string) {
	w.Header().Add("Access-Control-Allow-Credentials", "true")
	w.Header().Add("Access-Control-Allow-Origin", allow_origin)
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET,OPTIONS,POST,PUT,PATCH")
}
