package main

import (
	"github.com/go-martini/martini"

	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setup() *martini.ClassicMartini {
	os.Setenv("AUTH_USER", "default")
	os.Setenv("AUTH_PASS", "default")
	m := App()

	return m
}

func TestCatalog(t *testing.T) {
	m := setup()
	url := "/v2/catalog"

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", url, nil)
	m.ServeHTTP(res, req)

	// Without auth
	if res.Code != http.StatusUnauthorized {
		t.Error(url, "without auth should return 401")
	}

	res = httptest.NewRecorder()
	req.SetBasicAuth("default", "default")
	m.ServeHTTP(res, req)

	// With auth
	if res.Code != http.StatusOK {
		t.Error(url, "with auth should return 200 and it returned", res.Code)
	}

	// Is it a valid JSON?
	validJson(res.Body.Bytes(), url, t)

}

func TestCreateInstance(t *testing.T) {
	m := setup()
	url := "/v2/service_instances/the_instance"

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", url, nil)
	req.SetBasicAuth("default", "default")
	m.ServeHTTP(res, req)

	// With auth
	if res.Code != http.StatusCreated {
		t.Error(url, "with auth should return 201 and it returned", res.Code)
	}

	// Is it a valid JSON?
	validJson(res.Body.Bytes(), url, t)

	// Is it an empty object?
	if string(res.Body.Bytes()) != "{}" {
		t.Error(url, "should return an empty JSON")
	}
}

func TestBindInstance(t *testing.T) {
	m := setup()
	url := "/v2/service_instances/the_instance/service_bindings/the_binding"

	res := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", url, nil)
	req.SetBasicAuth("default", "default")
	m.ServeHTTP(res, req)

	// With auth
	if res.Code != http.StatusCreated {
		t.Error(url, "with auth should return 201 and it returned", res.Code)
	}

	// Is it a valid JSON?
	validJson(res.Body.Bytes(), url, t)

	// Does it contain "crendentials"
	if !strings.Contains(string(res.Body.Bytes()), "credentials") {
		t.Error(url, "should return credentials")
	}
}

func TestDeletes(t *testing.T) {
	m := setup()
	urls := []string{
		"/v2/service_instances/the_instance/service_bindings/the_binding",
		"/v2/service_instances/the_instance",
	}

	for _, url := range urls {
		res := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", url, nil)
		req.SetBasicAuth("default", "default")
		m.ServeHTTP(res, req)

		if res.Code != http.StatusOK {
			t.Error(url, "with auth should return 200 and it returned", res.Code)
		}

		// Is it a valid JSON?
		validJson(res.Body.Bytes(), url, t)

		// Is it an empty object?
		if string(res.Body.Bytes()) != "{}" {
			t.Error(url, "should return an empty JSON")
		}
	}
}

func validJson(response []byte, url string, t *testing.T) {
	var aJson map[string]interface{}
	if json.Unmarshal(response, &aJson) != nil {
		t.Error(url, "should return a valid json")
	}

}
