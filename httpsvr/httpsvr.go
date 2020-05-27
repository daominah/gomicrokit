// Package httpsvr supports http method, url params and
// logs all pairs of request/response. API is similar to standard net/http
package httpsvr

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/daominah/gomicrokit/gofast"
	"github.com/daominah/gomicrokit/log"
	"github.com/julienschmidt/httprouter"
)

// Server must be inited by calling func NewServer.
// Example usage in `a_examples/httpsvr/httpsvr.go`
type Server struct {
	// config defines parameters for running an HTTP server,
	// usually user should set ReadHeaderTimeout, ReadTimeout, WriteTimeout,
	// ReadTimeout and WriteTimeout should be bigger for a file server
	config *http.Server
	// router is a better http_ServeMux that supports http method, url params,
	// example: router.AddHandler("GET", "/match/:id", func(w,r))
	router *httprouter.Router
	// default NewServer set isEnableLog = true
	isEnableLog bool
	// default NewServer set isEnableMetric = true
	isEnableMetric bool
	metric         Metric
}

// NewServer returns a inited Server,
// for more configs, use NewServerWithConf instead of this func
func NewServer() *Server {
	return &Server{
		config: &http.Server{
			ReadHeaderTimeout: 20 * time.Second,
			ReadTimeout:       10 * time.Minute,
			WriteTimeout:      20 * time.Minute,
		},
		isEnableLog:    true,
		isEnableMetric: true,
		router:         httprouter.New(),
		metric:         NewMemoryMetric(),
	}
}

// NewServerWithConf returns a inited Server from input args,
// for simple usage, use NewServer instead of this func
func NewServerWithConf(config *http.Server, isEnableLog bool,
	isEnableMetric bool, metric Metric) *Server {
	if isEnableMetric && metric == nil {
		metric = NewMemoryMetric()
	}
	return &Server{
		config:         config,
		isEnableLog:    isEnableLog,
		isEnableMetric: isEnableMetric,
		router:         httprouter.New(),
		metric:         metric,
	}
}

// AddHandler must be called before ListenAndServe,
// ex: AddHandler("GET", "/", ExampleHandler())
func (s *Server) AddHandler(method string, path string, handler http.HandlerFunc) {
	defer func() { // example: add a same handler twice
		if r := recover(); r != nil {
			log.Infof("error when AddHandler: %v", r)
		}
	}()
	handlerWithLog := func(w http.ResponseWriter, r *http.Request) {
		requestId := gofast.GenUUID()
		ctx := context.WithValue(r.Context(), CtxRequestId, requestId)
		query := r.URL.Query().Encode()
		if query != "" {
			query = "?" + query
		}
		log.Condf(s.isEnableLog, "http request %v from %v: %v %v%v",
			requestId, r.RemoteAddr, r.Method, r.URL.Path, query)
		handler(w, r.WithContext(ctx))
	}
	if !s.isEnableMetric {
		s.router.HandlerFunc(method, path, handlerWithLog)
		return
	}
	metricKey := fmt.Sprintf("%v_%v", path, method)
	handlerWithMetric := func(w http.ResponseWriter, r *http.Request) {
		s.metric.Count(metricKey, 1)
		beginTime := time.Now()
		handlerWithLog(w, r)
		s.metric.Duration(metricKey, time.Since(beginTime))
	}
	s.router.HandlerFunc(method, path, handlerWithMetric)
}

// ListenAndServe listens on the TCP network address addr.
// Accepted connections are configured to enable TCP keep-alives.
func (s *Server) ListenAndServe(addr string) error {
	log.Infof("http server is listening on port %v", addr)
	s.config.Addr = addr
	err := s.config.ListenAndServe()
	return err
}

// ListenAndServe listens on the port s_config_Addr
func (s *Server) ListenAndServe2() error {
	return s.ListenAndServe(s.config.Addr)
}

// avoid context key conflict
type ctxKeyType string

// CtxRequestId is a internal request id
const CtxRequestId ctxKeyType = "CtxRequestId"

// WriteJson includes logging, input r is the corresponding request of the response
func (s Server) WriteJson(w http.ResponseWriter, r *http.Request, obj interface{}) (
	int, error) {
	bodyB, err := json.Marshal(obj)
	if err != nil {
		log.Condf(s.isEnableLog, "error when http respond %v: %v",
			GetRequestId(r), err)
		http.Error(w, err.Error(), 500)
		return 0, err
	}
	n, err := w.Write(bodyB)
	if err != nil {
		log.Condf(s.isEnableLog, "error when http respond %v: %v",
			GetRequestId(r), err)
		return n, err
	}
	log.Condf(s.isEnableLog, "http respond %v: %s", GetRequestId(r), bodyB)
	return n, nil
}

// ReadJson reads http request body to outPtr
func (s Server) ReadJson(r *http.Request, outPtr interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	log.Condf(s.isEnableLog, "http request %v body: %s", GetRequestId(r), body)
	err = json.Unmarshal(body, outPtr)
	return err
}

// GetUrlParams returns URL parameters from a http request as a map,
// ex: path `/match/:id` has param `id`
func GetUrlParams(r *http.Request) map[string]string {
	params := httprouter.ParamsFromContext(r.Context())
	result := make(map[string]string, len(params))
	if len(params) == 0 {
		return result
	}
	for _, param := range params {
		result[param.Key] = param.Value
	}
	return result
}

// GetRequestId returns the auto generated requestId
func GetRequestId(r *http.Request) string {
	return fmt.Sprintf("%v", r.Context().Value(CtxRequestId))
}

// ExampleHandler _
func ExampleHandler() http.HandlerFunc {
	// thing := initHandler() // one-time per-handler initialisation
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct{ Field0 string }
		_ = Server{isEnableLog: true}.ReadJson(r, &request)
		_ = GetUrlParams(r)
		Server{isEnableLog: true}.WriteJson(
			w, r, map[string]string{"Error": "", "Data": "PONG"})
	}
}

// ExampleHandlerError _
func ExampleHandlerError() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// not marshallable data
		Server{isEnableLog: true}.WriteJson(
			w, r, map[string]interface{}{"Data": func() {}})
	}
}
