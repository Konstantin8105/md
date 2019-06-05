package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func Test(t *testing.T) {

	tcs := []struct {
		handler        func(w http.ResponseWriter, r *http.Request)
		url            string
		expectFilename string
	}{
		{
			handler:        mainHandler,
			url:            "/",
			expectFilename: "test.main-index",
		},
		{
			handler:        articleHandler,
			url:            "/article/",
			expectFilename: "test.article-empty",
		},
		{
			handler:        articleHandler,
			url:            "/article/README.md",
			expectFilename: "test.article-README",
		},
		{
			handler:        articleHandler,
			url:            "/article/not_exist_file",
			expectFilename: "test.article-not-exist-file",
		},
	}

	for i := range tcs {
		t.Run(tcs[i].url, func(t *testing.T) {
			req := httptest.NewRequest("GET", tcs[i].url, nil)
			w := httptest.NewRecorder()
			tcs[i].handler(w, req)

			resp := w.Result()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			body = bytes.Replace(body, []byte("%2F"), []byte("/"), -1)
			body = bytes.Replace(body, []byte("+"), []byte(" "), -1)

			if os.Getenv("UPDATE") == "true" {
				err = ioutil.WriteFile(tcs[i].expectFilename, body, 0644)
				if err != nil {
					t.Fatal(err)
				}
			}

			content, err := ioutil.ReadFile(tcs[i].expectFilename)
			if err != nil {
				t.Fatal(err)
			}

			if !bytes.Equal(body, content) {
				t.Errorf("%s", body)
			}
		})
	}
}
