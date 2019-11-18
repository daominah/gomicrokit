package httpsvr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/daominah/gomicrokit/log"
	"github.com/daominah/gomicrokit/maths"
	"github.com/julienschmidt/httprouter"
)

// Server should be constructed by calling func NewServer, then be embedded.
// Example usage in `_examples/httpsvr/httpsvr.go`
type Server struct {
	router *httprouter.Router
	// TODO: add instrument metric for handlers
}

func NewServer() *Server { return &Server{router: httprouter.New()} }

// ex: AddHandler("GET", "/", ExampleHandler())
// all AddHandlers must call before ListenAndServe
func (s *Server) AddHandler(method string, path string, handler http.HandlerFunc) {
	defer func() {
		if r := recover(); r != nil {
			log.Infof("error when AddHandler: %v", r)
		}
	}()
	s.router.HandlerFunc(method, path, handler)
}

// ListenAndServe listens on the TCP network address addr.
// Accepted connections are configured to enable TCP keep-alives.
func (s Server) ListenAndServe(addr string) error {
	loggerWrapper := httpLogger{handler: s.router}
	log.Infof("http server is listening on port %v", addr)
	err := http.ListenAndServe(addr, loggerWrapper)
	return err
}

// httpLogger is a wrapper that help to log all requests
type httpLogger struct {
	handler http.Handler
}

const ctxRequestId = "ctxRequestId"

// log on every received request
func (l httpLogger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// http request's body can only be read once,
	// below codes help to read the body twice, the first read is for logging
	reqBodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body.Close()
	r.Body = ioutil.NopCloser(bytes.NewBuffer(reqBodyBytes))
	query := r.URL.Query().Encode()
	if query != "" {
		query = "?" + query
	}

	requestId := maths.GenUUID()[:8]
	log.Infof("request %v from %v: %v %v%v %v",
		requestId, r.RemoteAddr, r.Method, r.URL.Path, query, string(reqBodyBytes))

	ctx := context.WithValue(r.Context(), ctxRequestId, requestId)
	l.handler.ServeHTTP(w, r.WithContext(ctx))
}

// WriteJson includes logging,
// input r is the corresponding request of the response
func WriteJson(w http.ResponseWriter, r *http.Request, obj interface{}) {
	bodyB, err := json.Marshal(obj)
	if err != nil {
		errMsg := fmt.Sprintf("%v, obj: %#v", err, obj)
		WriteErr(w, r, http.StatusInternalServerError, errMsg)
		return
	}
	_, err = w.Write(bodyB)
	bodyS := string(bodyB)
	if err != nil {
		errMsg := fmt.Sprintf("error when writer write: %v, %v", err, bodyS)
		WriteErr(w, r, http.StatusInternalServerError, errMsg)
		return
	}
	requestId := r.Context().Value(ctxRequestId)
	log.Infof("respond %v successfully: %v", requestId, bodyS)
}

func WriteErr(w http.ResponseWriter, r *http.Request, code int, err string) {
	requestId := r.Context().Value(ctxRequestId)
	log.Infof("respond %v: code: %v, error: %v", requestId, code, err)
	http.Error(w, err, code)
}

// ReadJson reads http request body to outPtr
func ReadJson(r *http.Request, outPtr interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(body, outPtr)
	return err
}

func ExampleHandler() http.HandlerFunc {
	// thing := initHandler() // one-time per-handler initialisation
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct{ Field0 string }
		_ = ReadJson(r, &request)
		WriteJson(w, r, map[string]string{"Error": "", "Data": "PONG"})
	}
}

func ExampleHandlerError() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// not marshallable data
		WriteJson(w, r, map[string]interface{}{"Data": func() {}})
	}
}
