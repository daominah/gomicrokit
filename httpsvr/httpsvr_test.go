package httpsvr

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHttpSvr(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	ExampleHandler()(w, r)
	resBody, _ := ioutil.ReadAll(w.Result().Body)
	var obj struct {
		Error string
		Data  string
	}
	json.Unmarshal(resBody, &obj)
	if obj.Data != "PONG" {
		t.Error(obj)
	}

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest("GET", "/", nil)
	ExampleHandlerError()(w1, r1)
	if w1.Result().StatusCode != http.StatusInternalServerError {
		t.Error(w1.Result().Status)
	}
}
