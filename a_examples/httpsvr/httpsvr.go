package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/daominah/gomicrokit/httpsvr"
	"github.com/daominah/gomicrokit/log"
)

type Server struct {
	*httpsvr.Server
}

func (s *Server) Route() {
	s.AddHandler("GET", "/", s.index())
	// error a handle is already registered for path '/'
	s.AddHandler("GET", "/", s.index())
	s.AddHandler("GET", "/admin", s.auth(s.hello()))
	s.AddHandler("GET", "/error", httpsvr.ExampleHandlerError())
	s.AddHandler("GET", "/exception", s.exception())
}

func (s *Server) index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsvr.WriteJson(w, r, map[string]string{"Data": "Index page"})
	}
}

func (s *Server) hello() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpsvr.WriteJson(w, r, map[string]string{
			"Data": fmt.Sprintf("Hello %v", r.Context().Value("user")),
		})
	}
}

func (s Server) auth(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bearerAuth := r.Header.Get("Authorization")
		words := strings.Split(bearerAuth, " ")
		if len(words) != 2 || words[0] != "Bearer" {
			httpsvr.WriteErr(w, r, http.StatusUnauthorized,
				"need header Authorization: Bearer {token}")
			return
		}
		userName := words[1]
		ctx := context.WithValue(r.Context(), "user", userName)
		handle(w, r.WithContext(ctx))
	}
}

func (s *Server) exception() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var b *float64
		a := 1 / *b
		httpsvr.WriteJson(w, r, map[string]float64{"a": a})
	}
}

func main() {
	s := Server{Server: httpsvr.NewServer()}
	s.Route()
	port := ":8000"
	log.Infof("url0: http://127.0.0.1%v/", port)
	log.Infof("url1: http://127.0.0.1%v/admin", port)
	log.Infof("url2: http://127.0.0.1%v/error", port)
	err := s.ListenAndServe(port)
	if err != nil {
		log.Fatalf("error when s_ListenAndServe: %v", err)
	}
}
