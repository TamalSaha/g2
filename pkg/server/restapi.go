package server

import (
	"net/http"

	"github.com/appscode/pat"
)

func registerAPIHandlers(s *Server) {
	m := pat.New()

	m.Get("/jobs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		e := &event{tp: ctrlGetJob, result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	m.Get("/jobs/:handle", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params, _ := pat.FromContext(r.Context())
		e := &event{tp: ctrlGetJob, handle: params.Get(":handle"), result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	m.Get("/workers", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		e := &event{tp: ctrlGetWorker, args: &Tuple{}, result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	m.Get("/workers/:cando", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params, _ := pat.FromContext(r.Context())
		e := &event{tp: ctrlGetWorker, args: &Tuple{t0: params.Get(":cando")}, result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	m.Get("/cronjobs", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		e := &event{tp: ctrlGetCronJob, result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	//get job information using job handle
	m.Get("/cronjobs/:handle", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		params, _ := pat.FromContext(r.Context())
		e := &event{tp: ctrlGetCronJob, handle: params.Get(":handle"), result: createResCh()}
		s.ctrlEvtCh <- e
		res := <-e.result
		w.Write([]byte(res.(string)))
	}))

	http.Handle("/", m)
}
