package server

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/iostrovok/aura-test/storage"
)

// Start is an entry point for HTTP server.
func Start(ctx context.Context) {
	keeper := storage.New(ctx)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthcheck", healthCheck)
	mux.HandleFunc("/", initSessionsHandlers(keeper))

	logrus.Infof("HTTP SERVER is starting...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		logrus.Error(err.Error())
	}
}

// initSessionsHandlers provides function for "/sessions" path.
func initSessionsHandlers(keeper *storage.Storage) func(w http.ResponseWriter, req *http.Request) {
	logrus.Infof("HTTP SERVER is making handlers...")

	return func(w http.ResponseWriter, req *http.Request) {
		// catch exceptions
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("%+v\n", r)
			}
		}()

		switch req.Method {
		case http.MethodPost: // create new session
			createSessionHandler(keeper, w, req)
		case http.MethodGet: // list of all session
			listSessionsHandler(keeper, w, req)
		case http.MethodPut: // extend the session
			extendHandler(keeper, w, req)
		case http.MethodDelete: // Destroy the session
			destroyHandler(keeper, w, req)
		default:
			errorMethodRequest(w, req)
		}
	}
}

// errorMethodRequest is helper. It returns error about wrong HTTP method.
func errorMethodRequest(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
	if _, err := w.Write([]byte("Method " + req.Method + " is not allowed")); err != nil {
		logrus.Error(err.Error())
	}
}
