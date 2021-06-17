package server

import (
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"

	"github.com/iostrovok/aura-test/response"
	"github.com/iostrovok/aura-test/storage"
)

const (
	MaxAllowedExtendedTTL = int64(300)
	DefaultTTL            = int64(30)
	WrongPathError        = "wrong path"
	WrongTTLError         = "wrong TTL"
	WrongIDError          = "wrong session ID"
)

// createSession is interface method. It creates new session.
func createSessionHandler(keeper *storage.Storage, w http.ResponseWriter, req *http.Request) {
	/*
		create - Should take a TTL as an optional param, default should be 30 seconds.
		This API, when called, should return a unique session-id which should be UUID based.
		The session should then be stored in-memory.
		Only the sessions that have not expired are expected to be kept in memory.
		Any expired sessions should be removed automatically.
	*/
	// get POST parameter
	if err := req.ParseForm(); err != nil {
		logrus.Error(err.Error())
		jsonPrint(w, http.StatusBadRequest, response.Response{Error: err.Error()})

		return
	}

	ttl, err := strconv.ParseInt(req.FormValue("TTL"), 10, 64)
	if err != nil || ttl < 1 || ttl > DefaultTTL {
		ttl = DefaultTTL
	}

	// get new session uuid
	// always success
	id := keeper.Create(uint32(ttl))
	jsonPrint(w, http.StatusOK, response.Response{ID: id})
}

// extendHandler is interface method. It extends session ttl but no more then 300 sec.
func extendHandler(keeper *storage.Storage, w http.ResponseWriter, req *http.Request) {
	/*
		extend - Should take a mandatory session-id and an optional TTL param.
		When this API is called, if the session exists then it should be extended with the provided TTL
		or if the TTL is not provided then by 30 seconds.
		A 200 status code should be returned indicating success.
		If the session doesn't exist, then 404 should be returned.
		The max TTL allowed for this API is 300 seconds.
		Any greater value of TTL provided should be reduced to 300 seconds.
	*/

	id, ttl, err := parseURL(req)
	if err != nil {
		logrus.Error(err.Error())
		jsonPrint(w, http.StatusBadRequest, response.Response{Error: err.Error()})

		return
	}

	if findID := keeper.Extend(id, uint32(ttl)); !findID {
		logrus.Errorf("%s is not found", id)
		jsonPrint(w, http.StatusNotFound, response.Response{Error: "NotFound"})

		return
	}

	jsonPrint(w, http.StatusOK, response.Response{ID: id})
}

// destroyHandler is interface method. It deletes existing session and returns 404 is session id is not found.
func destroyHandler(keeper *storage.Storage, w http.ResponseWriter, req *http.Request) {
	/*
		Destroy - Should take a session-id as a mandatory param.
		When this API is called, if the session exists, then it should remove the session from its cache
		and return 200 status code, indicating success.
		If the session doesn't exist, then a 404 response should be returned.
	*/

	id, _, err := parseURL(req)
	if err != nil {
		logrus.Error(err.Error())
		jsonPrint(w, http.StatusBadRequest, response.Response{Error: err.Error()})

		return
	}

	status := http.StatusOK
	res := response.Response{ID: id}
	if find := keeper.Destroy(id); !find {
		logrus.Errorf("%s is not found", id)
		status = http.StatusNotFound
		res.Error = "NotFound"
	}

	jsonPrint(w, status, res)
}

// listSessionsHandler is interface method. It returns list of all active sessions and remaining TTL.
// Need to remember that some sessions may become expired during getting of data.
func listSessionsHandler(keeper *storage.Storage, w http.ResponseWriter, _ *http.Request) {
	/*
		list - Should just return a list of all the sessions that the service is currently tracking,
		each identified using its UUID and the corresponding TTL that is remaining.
	*/

	sessionJSONList := keeper.ListAllSessions()
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(sessionJSONList); err != nil {
		logrus.Error(err.Error())
	}
}

func healthCheck(w http.ResponseWriter, _ *http.Request) {
	jsonPrint(w, http.StatusOK, map[string]bool{"ok": true})
}

// jsonPrint is just helper.
func jsonPrint(w http.ResponseWriter, status int, data interface{}) {
	if b, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(data); err == nil {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(b); err != nil {
			logrus.Error(err.Error())
		}
	} else {
		logrus.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// OnlyNumbers checks on string has only digital.
var OnlyNumbers = regexp.MustCompile(`^\d+$`)

// parseURL is just helper. It's a wrapper over _parseURL.
func parseURL(req *http.Request) (string, int, error) {
	return _parseURL(req.URL.Path)
}

// parseURL is just helper.
// It returns id and ttl from url path if they are defined.
func _parseURL(url string) (string, int, error) {
	// POST, GET => /sessions
	// DELETE => /sessions/{id}
	// PUT => /sessions/{id}/{ttl}*
	in := strings.Split(strings.TrimRight(strings.TrimLeft(url, "/"), "/"), "/")

	if len(in) == 0 || in[0] != "sessions" {
		return "", 0, errors.New(WrongPathError)
	}

	id := ""
	if len(in) > 1 {
		if len(in[1]) != 36 {
			return "", 0, errors.New(WrongIDError)
		}
		id = in[1]
	}

	ttl := DefaultTTL
	if len(in) > 2 {
		// PUT => /sessions/{id}/{ttl}*  PUT
		if !OnlyNumbers.MatchString(in[2]) {
			return "", 0, errors.New(WrongTTLError)
		}

		var err error
		ttl, err = strconv.ParseInt(in[2], 10, 64)
		switch {
		case err != nil:
			return "", 0, errors.New(WrongTTLError)
		case ttl < 0:
			ttl = DefaultTTL
		case ttl > MaxAllowedExtendedTTL:
			ttl = MaxAllowedExtendedTTL
		}
	}

	return id, int(ttl), nil
}
