package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
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
			url:            "/article//////////////file_not_exist.md",
			expectFilename: "test.article-file-not-exist-md",
		},
		{
			handler:        articleHandler,
			url:            "/article/LICENSE",
			expectFilename: "test.article-LICENSE",
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

			body = bytes.Replace(body, []byte("\r"), []byte(""), -1)
			content = bytes.Replace(content, []byte("\r"), []byte(""), -1)

			if !bytes.Equal(body, content) {
				text := ShowDiff(string(content), string(body))
				t.Errorf("%s", text)
			}
		})
	}
}

// ShowDiff will print two strings vertically next to each other so that line
// differences are easier to read.
func ShowDiff(a, b string) string {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")
	maxLines := int(math.Max(float64(len(aLines)), float64(len(bLines))))
	out := "\n"

	for lineNumber := 0; lineNumber < maxLines; lineNumber++ {
		aLine := ""
		bLine := ""

		// Replace NULL characters with a dot. Otherwise the strings will look
		// exactly the same but have different length (and therfore not be
		// equal).
		if lineNumber < len(aLines) {
			aLine = strconv.Quote(aLines[lineNumber])
		}
		if lineNumber < len(bLines) {
			bLine = strconv.Quote(bLines[lineNumber])
		}

		diffFlag := " "
		if aLine != bLine {
			diffFlag = "*"
		}
		out += fmt.Sprintf("%s %3d %-40s%s\n", diffFlag, lineNumber+1, aLine, bLine)

		if lineNumber > len(aLines) || lineNumber > len(bLines) {
			out += "and more other ..."
			break
		}
	}

	return out
}
