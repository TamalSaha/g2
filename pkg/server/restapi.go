package server

import (
	"net/http"
	"time"

	"github.com/appscode/log"
	"github.com/go-macaron/toolbox"
	"gopkg.in/macaron.v1"
)

func getJob(s *Server, ctx *macaron.Context) string {
	e := &event{tp: ctrlGetJob,
		handle: ctx.Params("handle"), result: createResCh()}
	s.ctrlEvtCh <- e
	res := <-e.result

	return res.(string)
}

func getWorker(s *Server, ctx *macaron.Context) string {
	e := &event{tp: ctrlGetWorker,
		args: &Tuple{t0: ctx.Params("cando")}, result: createResCh()}
	s.ctrlEvtCh <- e
	res := <-e.result

	return res.(string)
}

func getCronJob(s *Server, ctx *macaron.Context) string {
	e := &event{tp: ctrlGetCronJob,
		handle: ctx.Params("handle"), result: createResCh()}
	s.ctrlEvtCh <- e
	res := <-e.result

	return res.(string)
}

func registerWebHandler(s *Server) {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())
	m.Use(toolbox.Toolboxer(m))

	m.Get("/jobs", func(ctx *macaron.Context) string {
		return getJob(s, ctx)
	})

	m.Get("/jobs/:handle", func(ctx *macaron.Context) string {
		return getJob(s, ctx)
	})

	m.Get("/workers", func(ctx *macaron.Context) string {
		return getWorker(s, ctx)
	})

	m.Get("/workers/:cando", func(ctx *macaron.Context) string {
		return getWorker(s, ctx)
	})

	m.Get("/cronjobs", func(ctx *macaron.Context) string {
		return getCronJob(s, ctx)
	})

	//get job information using job handle
	m.Get("/cronjobs/:handle", func(ctx *macaron.Context) string {
		return getCronJob(s, ctx)
	})

	log.Infof("listening on %s (%s)\n", s.config.RestAPIAddress, macaron.Env)
	srv := &http.Server{
		Addr:         s.config.RestAPIAddress,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      m,
	}
	log.Fatalln(srv.ListenAndServe())
}
