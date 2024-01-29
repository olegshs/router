package router

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func ExampleRouter_ParseMap() {
	r := DefaultRouter()

	r.ParseMap(
		map[string]interface{}{
			"GET /": "pages.main",
			"(group name)": map[string]interface{}{
				"$use": "test",
				"GET, POST /test/{id}": map[string]interface{}{
					"$name": "pages.test",
					"id":    `^\d+$`,
				},
			},
			"/api": map[string]interface{}{
				"$use": "test",
				"GET":  "api.index",
				"/articles": map[string]interface{}{
					"GET":  "articles.index",
					"POST": "articles.create",
					"/{id}": map[string]interface{}{
						"$where": map[string]interface{}{
							"id": `^\d+$`,
						},
						"GET":    "articles.get",
						"PUT":    "articles.update",
						"DELETE": "articles.delete",
					},
				},
			},
		},
		func(routeName string) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "Route:  %s\n", routeName)
				fmt.Fprintf(w, "Params: %v\n", ParamsFromRequest(r).Map())
			})
		},
		func(middlewareName string) MiddlewareFunc {
			switch middlewareName {
			case "test":
				return func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test", "OK")
						next.ServeHTTP(w, r)
					})
				}
			default:
				panic(fmt.Errorf("unknown middleware: %s", middlewareName))
			}
		},
	)
}

func ExampleRouter_Get() {
	r := DefaultRouter()

	r.Get("/hello").
		HandleFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, "Hello, world!")
		})
}

func ExampleRouter_Group() {
	r := DefaultRouter()

	r.Group(func(r *Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "OK")
				next.ServeHTTP(w, r)
			})
		})

		r.Get("/hello").
			HandleFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, "Hello, world!")
			})
	})
}

func ExampleRouter_Prefix() {
	r := DefaultRouter()

	r.Prefix("/articles/{id}", func(r *Router) {
		r.Where("id", regexp.MustCompile(`^\d+$`))

		r.Get("").
			HandleFunc(func(w http.ResponseWriter, r *http.Request) {
				id := ParamsFromRequest(r).ByName("id")
				fmt.Fprintf(w, "Get article #%s\n", id)
			})

		r.Get("/edit").
			HandleFunc(func(w http.ResponseWriter, r *http.Request) {
				id := ParamsFromRequest(r).ByName("id")
				fmt.Fprintf(w, "Edit article #%s\n", id)
			})
	})
}

func ExampleRouter_Url() {
	r := DefaultRouter()

	r.Get("/articles/{id}").
		Name("articles.get")

	url, err := r.Url("articles.get", 111)
	fmt.Println(url, err)
}

func ExampleRouter_HandleMethodNotAllowed() {
	r := DefaultRouter()

	r.HandleMethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, "Method %s is not allowed for %s\n", r.Method, r.URL)
	}))
}

func ExampleRouter_HandlePanic() {
	r := DefaultRouter()

	r.HandlePanic(func(w http.ResponseWriter, r *http.Request, e interface{}) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v\n", e)
	})
}

func TestRouter_ParseMap(t *testing.T) {
	r := New()

	r.ParseMap(
		map[string]interface{}{
			"GET /": "pages.main",
			"(group name)": map[string]interface{}{
				"$use": "test",
				"GET, POST /test/{id}": map[string]interface{}{
					"$name": "pages.test",
					"id":    `^\d+$`,
				},
			},
			"/api": map[string]interface{}{
				"$use": "test",
				"GET":  "api.index",
				"/articles": map[string]interface{}{
					"GET":  "articles.index",
					"POST": "articles.create",
					"/{id}": map[string]interface{}{
						"$where": map[string]interface{}{
							"id": `^\d+$`,
						},
						"GET":    "articles.get",
						"PUT":    "articles.update",
						"DELETE": "articles.delete",
					},
				},
			},
		},
		func(routeName string) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				params := ParamsFromRequest(r).Map()
				fmt.Fprintf(w, "route: %s, params: %v\n", routeName, params)
			})
		},
		func(middlewareName string) MiddlewareFunc {
			switch middlewareName {
			case "test":
				return func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test", "OK")
						next.ServeHTTP(w, r)
					})
				}
			default:
				panic(fmt.Errorf("unknown middleware: %s", middlewareName))
			}
		},
	)

	{
		resp := testRequest(r, http.MethodGet, "/", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertHeaderMissing(t, resp.Header, "X-Test")
		assertBody(t, resp.Body, "route: pages.main, params: map[]\n")
	}
	{
		a := [][3]string{
			{http.MethodGet, "/test/111", "route: pages.test, params: map[id:111]\n"},
			{http.MethodPost, "/test/111", "route: pages.test, params: map[id:111]\n"},
			{http.MethodGet, "/api", "route: api.index, params: map[]\n"},
			{http.MethodGet, "/api/articles", "route: articles.index, params: map[]\n"},
			{http.MethodPost, "/api/articles", "route: articles.create, params: map[]\n"},
			{http.MethodGet, "/api/articles/111", "route: articles.get, params: map[id:111]\n"},
			{http.MethodPut, "/api/articles/111", "route: articles.update, params: map[id:111]\n"},
			{http.MethodDelete, "/api/articles/111", "route: articles.delete, params: map[id:111]\n"},
		}
		for _, v := range a {
			resp := testRequest(r, v[0], v[1], nil, nil)
			assertStatus(t, resp.StatusCode, http.StatusOK)
			assertHeader(t, resp.Header, "X-Test", "OK")
			assertBody(t, resp.Body, v[2])
		}
	}
}

func TestRouter_Get(t *testing.T) {
	r := New()

	r.Get("/{id}").
		Where("id", regexp.MustCompile(`^\d+$`)).
		HandleFunc(func(w http.ResponseWriter, r *http.Request) {
			id := ParamsFromRequest(r).ByName("id")
			fmt.Fprintf(w, "get by id: %s\n", id)
		})

	r.Get("/{name}").
		HandleFunc(func(w http.ResponseWriter, r *http.Request) {
			name := ParamsFromRequest(r).ByName("name")
			fmt.Fprintf(w, "get by name: %s\n", name)
		})

	{
		resp := testRequest(r, http.MethodGet, "/111", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertBody(t, resp.Body, "get by id: 111\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/aaa", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertBody(t, resp.Body, "get by name: aaa\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/aaa/", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusMovedPermanently)
		assertHeader(t, resp.Header, "Location", "/aaa")
	}
}

func TestRouter_Post(t *testing.T) {
	r := New()

	r.Post("/test").
		HandleFunc(func(w http.ResponseWriter, r *http.Request) {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			value := r.Form.Get("value")
			fmt.Fprintln(w, value)
		})

	{
		resp := testRequest(r, http.MethodPost, "/test", nil, map[string]string{"value": "OK"})
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertBody(t, resp.Body, "OK\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/test", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusMethodNotAllowed)
	}
}

func TestRouter_Group(t *testing.T) {
	r := New()

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	m := MiddlewareFunc(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test", "OK")
			h.ServeHTTP(w, r)
		})
	})

	r.Get("/").
		Handle(h)

	r.Group(func(r *Router) {
		r.Use(m)

		r.Get("/test").
			Handle(h)
	})

	{
		resp := testRequest(r, http.MethodGet, "/", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertHeaderMissing(t, resp.Header, "X-Test")
		assertBody(t, resp.Body, "OK\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/test", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertHeader(t, resp.Header, "X-Test", "OK")
		assertBody(t, resp.Body, "OK\n")
	}
}

func TestRouter_Prefix(t *testing.T) {
	r := New()

	r.Prefix("/users", func(r *Router) {
		r.Use(func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test-A", "A")
				h.ServeHTTP(w, r)
			})
		})

		r.Prefix("/{userId}", func(r *Router) {
			r.Where("userId", regexp.MustCompile(`^\d+$`))

			r.Get("").
				HandleFunc(func(w http.ResponseWriter, r *http.Request) {
					params := ParamsFromRequest(r)
					userId := params.ByName("userId")
					fmt.Fprintf(w, "user: %s\n", userId)
				})

			r.Prefix("/articles", func(r *Router) {
				r.Use(func(h http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.Header().Set("X-Test-B", "B")
						h.ServeHTTP(w, r)
					})
				})

				r.Prefix("/{articleId}", func(r *Router) {
					r.Where("articleId", regexp.MustCompile(`^\d+$`))

					r.Get("").
						HandleFunc(func(w http.ResponseWriter, r *http.Request) {
							params := ParamsFromRequest(r)
							userId := params.ByName("userId")
							articleId := params.ByName("articleId")
							fmt.Fprintf(w, "user: %s, article: %s\n", userId, articleId)
						})
				})
			})
		})
	})

	{
		resp := testRequest(r, http.MethodGet, "/users/111", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertHeader(t, resp.Header, "X-Test-A", "A")
		assertHeaderMissing(t, resp.Header, "X-Test-B")
		assertBody(t, resp.Body, "user: 111\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/users/111/articles/222", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)
		assertHeader(t, resp.Header, "X-Test-A", "A")
		assertHeader(t, resp.Header, "X-Test-B", "B")
		assertBody(t, resp.Body, "user: 111, article: 222\n")
	}
	{
		resp := testRequest(r, http.MethodGet, "/users/aaa/articles/222", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusNotFound)
		assertHeader(t, resp.Header, "X-Test-A", "A")
		assertHeader(t, resp.Header, "X-Test-B", "B")
	}
	{
		resp := testRequest(r, http.MethodGet, "/users/111/articles/bbb", nil, nil)
		assertStatus(t, resp.StatusCode, http.StatusNotFound)
		assertHeader(t, resp.Header, "X-Test-A", "A")
		assertHeader(t, resp.Header, "X-Test-B", "B")
	}
}

func TestRouter_Url(t *testing.T) {
	r := New()

	r.Prefix("/users/{userId}", func(r *Router) {
		r.Where("userId", regexp.MustCompile(`^\d+$`))

		r.Get("/articles/{articleId}").
			Where("articleId", regexp.MustCompile(`^\d+$`)).
			Name("users.articles.get")
	})

	u, err := r.Url("users.articles.get", 111, 222)
	if err != nil {
		t.Fatal(err)
	}
	if u != "/users/111/articles/222" {
		t.Errorf("%s != %s", u, "/users/111/articles/222")
	}

	_, err = r.Url("users.articles.get", "aaa", "bbb")
	assertError(t, err, ErrInvalidParameter)
}

func testRequest(
	handler http.Handler, method string, target string, headers map[string]string, data map[string]string,
) *http.Response {
	var body io.Reader
	if len(data) > 0 {
		form := url.Values{}
		for k, v := range data {
			form.Add(k, v)
		}

		body = strings.NewReader(form.Encode())
	}

	r := httptest.NewRequest(method, target, body)

	if len(data) > 0 {
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		r.Header.Add(k, v)
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	return w.Result()
}

func assertStatus(t *testing.T, status, expected int) {
	t.Helper()

	if status != expected {
		t.Errorf(
			"status code: %d != %d",
			status, expected,
		)
	}
}

func assertBody(t *testing.T, body io.Reader, expected string) {
	t.Helper()

	b, err := ioutil.ReadAll(body)
	if err != nil {
		t.Errorf("body: %s", err)
		return
	}

	if string(b) != expected {
		t.Errorf(
			"body:\n"+
				"got:      %s\n"+
				"expected: %s",
			strconv.Quote(string(b)),
			strconv.Quote(expected),
		)
	}
}

func assertError(t *testing.T, err, expected error) {
	t.Helper()

	if !errors.Is(err, expected) {
		t.Errorf(
			"invalid error: %v\n"+
				"expected: %v",
			err,
			expected,
		)
	}
}

func assertHeaderMissing(t *testing.T, header http.Header, key string) {
	t.Helper()

	if _, ok := header[key]; ok {
		t.Errorf(
			"unexpected header: %s",
			key,
		)
	}
}

func assertHeader(t *testing.T, header http.Header, key string, expected string) {
	t.Helper()

	if _, ok := header[key]; !ok {
		t.Errorf(
			"header missing: %s",
			key,
		)
		return
	}

	value := header.Get(key)
	if value != expected {
		t.Errorf(
			"header %s:\n"+
				"got:      %s\n"+
				"expected: %s\n",
			key,
			value,
			expected,
		)
	}
}
