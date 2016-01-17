package rux_test

import (
	"net/http"
	"testing"

	"github.com/cswank/rux"
)

func BenchmarkMux(b *testing.B) {
	r := rux.New("bench")
	handler := func(w http.ResponseWriter, r *http.Request) {}
	r.Get("/v1/{v1}", handler)

	request, _ := http.NewRequest("GET", "/v1/anything", nil)
	for i := 0; i < b.N; i++ {
		r.ServeHTTP(nil, request)
	}
}
