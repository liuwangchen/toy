package httprpc

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/liuwangchen/toy/pkg/host"
)

const (
	appJSONStr = "application/json"
	address    = ":12323"
)

type User struct {
	Name string `json:"name"`
}

func corsFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			log.Println("cors:", r.Method, r.RequestURI)
			w.Header().Set("Access-Control-Allow-Methods", r.Method)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func authFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println("auth:", r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func loggingFilter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println("logging:", r.Method, r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func TestRoute(t *testing.T) {
	ctx := context.Background()
	server, err := NewServerConn(
		WithServerFilter(corsFilter, loggingFilter),
		//Address(address),
	)
	if err != nil {
		t.Fatal(err)
	}
	route := server.Route("/v1")
	route.GET("/users/{name}", func(ctx Context) error {
		u := new(User)
		u.Name = ctx.Vars().Get("name")
		return ctx.Result(200, u)
	}, authFilter)
	route.POST("/users", func(ctx Context) error {
		u := new(User)
		if err := ctx.Bind(u); err != nil {
			return err
		}
		return ctx.Result(201, u)
	})
	route.PUT("/users", func(ctx Context) error {
		u := new(User)
		if err := ctx.Bind(u); err != nil {
			return err
		}
		h := ctx.Middleware(func(ctx context.Context, in interface{}) (interface{}, error) {
			return u, nil
		})
		return ctx.Returns(h(ctx, u))
	})

	if e, err := server.Endpoint(); err != nil || e == nil {
		t.Fatal(e, err)
	}
	go func() {
		if err := server.Start(ctx); err != nil {
			panic(err)
		}
	}()
	time.Sleep(time.Second)
	testRoute(t, server)
	_ = server.Stop(ctx)
}

func testRoute(t *testing.T, srv *ServerConn) {
	port, ok := host.Port(srv.lis)
	if !ok {
		t.Fatalf("extract port error: %v", srv.lis)
	}
	base := fmt.Sprintf("http://127.0.0.1:%d/v1", port)
	// GET
	resp, err := http.Get(base + "/users/foo")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("code: %d", resp.StatusCode)
	}
	if v := resp.Header.Get("Content-Type"); v != appJSONStr {
		t.Fatalf("contentType: %s", v)
	}
	u := new(User)
	if err = json.NewDecoder(resp.Body).Decode(u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "foo" {
		t.Fatalf("got %s want foo", u.Name)
	}
	// POST
	resp, err = http.Post(base+"/users", appJSONStr, strings.NewReader(`{"name":"bar"}`))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 201 {
		t.Fatalf("code: %d", resp.StatusCode)
	}
	if v := resp.Header.Get("Content-Type"); v != appJSONStr {
		t.Fatalf("contentType: %s", v)
	}
	u = new(User)
	if err = json.NewDecoder(resp.Body).Decode(u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "bar" {
		t.Fatalf("got %s want bar", u.Name)
	}
	// PUT
	req, _ := http.NewRequest("PUT", base+"/users", strings.NewReader(`{"name":"bar"}`))
	req.Header.Set("Content-Type", appJSONStr)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("code: %d", resp.StatusCode)
	}
	if v := resp.Header.Get("Content-Type"); v != appJSONStr {
		t.Fatalf("contentType: %s", v)
	}
	u = new(User)
	if err = json.NewDecoder(resp.Body).Decode(u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "bar" {
		t.Fatalf("got %s want bar", u.Name)
	}
	// OPTIONS
	req, _ = http.NewRequest("OPTIONS", base+"/users", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("code: %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Methods") != "OPTIONS" {
		t.Fatal("cors failed")
	}
}

func TestRouter_Group(t *testing.T) {
	r := &Router{}
	rr := r.Group("a", func(http.Handler) http.Handler { return nil })
	if !reflect.DeepEqual("a", rr.prefix) {
		t.Errorf("expected %q, got %q", "a", rr.prefix)
	}
}

func TestHandle(t *testing.T) {
	//server, err := NewServerConn(Address(address))
	server, err := NewServerConn()
	if err != nil {
		t.Fatal(err)
	}
	r := newRouter("/", server)
	h := func(i Context) error {
		return nil
	}
	r.GET("/get", h)
	r.HEAD("/head", h)
	r.PATCH("/patch", h)
	r.DELETE("/delete", h)
	r.CONNECT("/connect", h)
	r.OPTIONS("/options", h)
	r.TRACE("/trace", h)
}
