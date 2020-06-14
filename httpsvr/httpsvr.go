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
	"github.com/daominah/gomicrokit/metric"
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
	Metric         metric.Metric
}

// NewServer returns a inited Server,
// for more configs, use NewServerWithConf instead of this func
func NewServer() *Server {
	router := httprouter.New()
	config := NewDefaultConfig()
	config.Handler = router
	return &Server{
		config:         config,
		isEnableLog:    true,
		isEnableMetric: true,
		router:         router,
		Metric:         metric.NewMemoryMetric(),
	}
}

// NewServerWithConf returns a inited Server from input args.
// This func will ignore config_Handler, you have to use Server_AddHandler
// to define the router.
// For simple usage, use NewServer instead of this func.
func NewServerWithConf(config *http.Server, isEnableLog bool,
	isEnableMetric bool, metric0 metric.Metric) *Server {
	if isEnableMetric && metric0 == nil {
		metric0 = metric.NewMemoryMetric()
	}
	if config == nil {
		config = NewDefaultConfig()
	}
	router := httprouter.New()
	config.Handler = router
	return &Server{
		config:         config,
		isEnableLog:    isEnableLog,
		isEnableMetric: isEnableMetric,
		router:         router,
		Metric:         metric0,
	}
}

// AddHandler must be called before ListenAndServe,
// ex: AddHandler("GET", "/", ExampleHandler()).
func (s *Server) AddHandler(method string, path string, handler http.HandlerFunc) {
	defer func() { // in case of adding a same handler twice
		if r := recover(); r != nil {
			log.Infof("error when AddHandler: %v", r)
		}
	}()
	// be careful with augmenting handler, example stack overflow:
	// 	f := func() { log.Println("f called") }
	//	f = func() { f() }
	//	f()

	var augmented1 http.HandlerFunc
	if !s.isEnableMetric {
		augmented1 = handler
	} else {
		metricKey := fmt.Sprintf("%v_%v", path, method)
		augmented1 = func(w http.ResponseWriter, r *http.Request) {
			s.Metric.Count(metricKey)
			beginTime := time.Now()
			handler(w, r)
			s.Metric.Duration(metricKey, time.Since(beginTime))
		}
	}

	var augmented2 http.HandlerFunc
	if !s.isEnableLog {
		augmented2 = augmented1
	} else {
		augmented2 = func(w http.ResponseWriter, r *http.Request) {
			requestId := gofast.GenUUID()
			ctx := context.WithValue(r.Context(), CtxRequestId, requestId)
			query := r.URL.Query().Encode()
			if query != "" {
				query = "?" + query
			}
			log.Condf(s.isEnableLog, "http request %v from %v: %v %v%v",
				requestId, r.RemoteAddr, r.Method, r.URL.Path, query)
			augmented1(w, r.WithContext(ctx))
			log.Condf(s.isEnableLog, "http responded %v to %v: %v %v%v",
				requestId, r.RemoteAddr, r.Method, r.URL.Path, query)
		}
	}

	s.router.HandlerFunc(method, path, augmented2)
}

// ListenAndServe listens on the TCP network address addr.
// Accepted connections are configured to enable TCP keep-alives.
func (s *Server) ListenAndServe(addr string) error {
	s.config.Addr = addr
	return s.config.ListenAndServe()
}

// ListenAndServe listens on the port s_config_Addr
func (s *Server) ListenAndServe2() error {
	return s.ListenAndServe(s.config.Addr)
}

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
	w.Header().Set("Content-Type", "application/json")
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
	log.Condf(s.isEnableLog, "http request body %v: %s", GetRequestId(r), body)
	err = json.Unmarshal(body, outPtr)
	return err
}

var emptyServer = &Server{isEnableLog: true}

func WriteJson(w http.ResponseWriter, r *http.Request, obj interface{}) (int, error) {
	return emptyServer.WriteJson(w, r, obj)
}

func ReadJson(r *http.Request, outPtr interface{}) error {
	return emptyServer.ReadJson(r, outPtr)
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

// ctxKeyType is used for avoiding context key conflict
type ctxKeyType string

// CtxRequestId is a internal request id
const CtxRequestId ctxKeyType = "CtxRequestId"

// GetRequestId returns the auto generated unique requestId
func GetRequestId(r *http.Request) string {
	return fmt.Sprintf("%v", r.Context().Value(CtxRequestId))
}

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

func ExampleHandlerError() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// not marshallable data
		Server{isEnableLog: true}.WriteJson(
			w, r, map[string]interface{}{"Data": func() {}})
	}
}

// NewDefaultConfig is my suggestion of a http server config,
// feel free to modified base on your circumstance
func NewDefaultConfig() *http.Server {
	return &http.Server{
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      20 * time.Minute,
	}
}
