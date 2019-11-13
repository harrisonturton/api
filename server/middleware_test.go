package server

import (
	"bytes"
	//`ZZZZ	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	mockUrl = "http://localhost"
)

func TestPreflightHeaders(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	}
	req := httptest.NewRequest(http.MethodOptions, mockUrl, nil)
	w := httptest.NewRecorder()
	handler(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf(
			"%s request failed: expected status code %d but got %d",
			http.MethodOptions,
			http.StatusOK,
			resp.StatusCode,
		)
	}
}
