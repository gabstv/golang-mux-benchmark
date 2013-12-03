package mux_bench_test

import (
	"github.com/gocraft/web"
	"github.com/gorilla/mux"
	"github.com/codegangsta/martini"
	"testing"
	"fmt"
	"net/http"
	"net/http/httptest"
	"crypto/sha1"
	"io"
)

type RouterBuilder func(namespaces []string, resources []string) http.Handler

//
// Types / Methods needed by gocraft/web:
//
type BenchContext struct{}
func (c *BenchContext) Action(w web.ResponseWriter, r *web.Request) {
	fmt.Fprintf(w, "hello")
}


//
// Benchmarks for gocraft/web:
//
func BenchmarkGocraftWebSimple(b *testing.B) {
	router := web.New(BenchContext{})
	router.Get("/action",func(rw web.ResponseWriter, r *web.Request) {
		fmt.Fprintf(rw, "hello")
	})
	
	rw, req := testRequest("GET", "/action")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rw, req)
	}
}

func BenchmarkGocraftWebRoute15(b *testing.B) {
	benchmarkRoutesN(b, 1, gocraftWebRouterFor)
}

func BenchmarkGocraftWebRoute75(b *testing.B) {
	benchmarkRoutesN(b, 5, gocraftWebRouterFor)
}

func BenchmarkGocraftWebRoute150(b *testing.B) {
	benchmarkRoutesN(b, 10, gocraftWebRouterFor)
}

func BenchmarkGocraftWebRoute300(b *testing.B) {
	benchmarkRoutesN(b, 20, gocraftWebRouterFor)
}

func BenchmarkGocraftWebRoute3000(b *testing.B) {
	benchmarkRoutesN(b, 200, gocraftWebRouterFor)
}

//
// Benchmarks for gorilla/mux:
//
func BenchmarkGorillaMuxSimple(b *testing.B) {
	router := mux.NewRouter()
	router.HandleFunc("/action", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello")
	}).Methods("GET")
	
	rw, req := testRequest("GET", "/action")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router.ServeHTTP(rw, req)
	}
}

// func BenchmarkGorillaMuxRoute15(b *testing.B) {
// 	const N = 1
// 	namespaces, resources, requests := resourceSetup(N)
// 	router := gorillaMuxRouterFor(namespaces, resources)
// 	benchmarkRoutes(b, router, requests)
// }


//
// Benchmarks for codegangsta/martini:
//
func BenchmarkCodegangstaMartiniSimple(b *testing.B) {
	r := martini.NewRouter()
	m := martini.New()
	m.Action(r.Handle)
	
	r.Get("/action", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "hello")
	})
	
	rw, req := testRequest("GET", "/action")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ServeHTTP(rw, req)
	}
}

//
// Helpers:
//
func testRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()

	return recorder, request
}

func benchmarkRoutesN(b *testing.B, N int, builder RouterBuilder) {
	namespaces, resources, requests := resourceSetup(N)
	router := builder(namespaces, resources)
	benchmarkRoutes(b, router, requests)
}

// Returns a routeset with N *resources per namespace*. so N=1 gives about 15 routes
func resourceSetup(N int) (namespaces []string, resources []string, requests []*http.Request) {
	namespaces = []string{"admin", "api", "site"}
	resources = []string{}
	
	for i := 0; i < N; i += 1 {
		sha1 := sha1.New()
		io.WriteString(sha1, fmt.Sprintf("%d", i))
		strResource := fmt.Sprintf("%x", sha1.Sum(nil))
		resources = append(resources, strResource)
	}
	
	for _, ns := range namespaces {
		for _, res := range resources {
			req, _ := http.NewRequest("GET", "/"+ns+"/"+res, nil)
			requests = append(requests, req)
			req, _ = http.NewRequest("POST", "/"+ns+"/"+res, nil)
			requests = append(requests, req)
			req, _ = http.NewRequest("GET", "/"+ns+"/"+res+"/3937", nil)
			requests = append(requests, req)
			req, _ = http.NewRequest("PUT", "/"+ns+"/"+res+"/3937", nil)
			requests = append(requests, req)
			req, _ = http.NewRequest("DELETE", "/"+ns+"/"+res+"/3937", nil)
			requests = append(requests, req)
		}
	}
	
	return
}

func gocraftWebRouterFor(namespaces []string, resources []string) http.Handler {
	router := web.New(BenchContext{})
	for _, ns := range namespaces {
		subrouter := router.Subrouter(BenchContext{}, "/"+ns)
		for _, res := range resources {
			subrouter.Get("/"+res, (*BenchContext).Action)
			subrouter.Post("/"+res, (*BenchContext).Action)
			subrouter.Get("/"+res+"/:id", (*BenchContext).Action)
			subrouter.Put("/"+res+"/:id", (*BenchContext).Action)
			subrouter.Delete("/"+res+"/:id", (*BenchContext).Action)
		}
	}
	return router
}

func benchmarkRoutes(b *testing.B, handler http.Handler, requests []*http.Request) {
	recorder := httptest.NewRecorder()
	reqId := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if reqId >= len(requests) {
			reqId = 0
		}
		req := requests[reqId]
		handler.ServeHTTP(recorder, req)

		// if recorder.Code != 200 {
		// panic("wat")
		// }

		reqId += 1
	}
}

