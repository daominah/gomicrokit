package httpsvr

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttp(t *testing.T) {
	s := NewServer()
	handler := s.router

	// handle 0
	s.AddHandler("GET", "/", ExampleHandler())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)
	resBody, _ := ioutil.ReadAll(w.Result().Body)
	var obj struct {
		Error string
		Data  string
	}
	json.Unmarshal(resBody, &obj)
	if obj.Data != "PONG" {
		t.Error(obj)
	}

	// handle 1
	s.AddHandler("GET", "/error", ExampleHandlerError())
	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/error", nil)
	handler.ServeHTTP(w1, r1)
	if w1.Result().StatusCode != http.StatusInternalServerError {
		t.Errorf("expected InternalServerError but %v", w1.Result().Status)
	}

	// handle 2
	type ParamQueryW struct {
		ParamId string
		ParamF2 string
		QueryQ1 string
		QueryQ2 string
	}
	s.AddHandler("GET", "/match/:id",
		func(w http.ResponseWriter, r *http.Request) {
			res := ParamQueryW{
				ParamId: GetUrlParams(r)["id"],
				QueryQ1: r.FormValue("q1"),
				QueryQ2: r.FormValue("q2"),
			}
			s.WriteJson(w, r, res)
		})
	w2 := httptest.NewRecorder()
	r2 := httptest.NewRequest("GET", "/match/119?q1=lan&q2=dt", nil)
	handler.ServeHTTP(w2, r2)
	bodyB, _ := ioutil.ReadAll(w2.Result().Body)
	var data ParamQueryW
	err := json.Unmarshal(bodyB, &data)
	if err != nil {
		t.Error(err, string(bodyB))
	}
	if data.ParamId != "119" || data.QueryQ1 != "lan" || data.QueryQ2 != "dt" {
		t.Errorf("data: %#v", data)
	}

	t.Log(s.Metric.GetCurrentMetric())
}
