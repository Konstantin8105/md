package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(mainHandler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	expect := []byte(`
<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			.markdown-body {
				box-sizing: border-box;
				min-width: 200px;
				max-width: 600px;
				margin: 0 auto;
				padding: 45px;
			}
			@media (max-width: 767px) {
				.markdown-body {
					padding: 15px;
				}
			}
	</style>
	</head>
	<body>
		<article class="markdown-body">
			<h1>List of articles:</h1>

<ul>
<li><p><a href="/articles/README.md">README.md</a></p></li>

<li><p><a href="/articles/testdata/folder with space/testSpace.md">testdata/folder with space/testSpace.md</a></p></li>

<li><p><a href="/articles/testdata/test.md">testdata/test.md</a></p></li>

<li><p><a href="/articles/vendor/github.com/russross/blackfriday/README.md">vendor/github.com/russross/blackfriday/README.md</a></p></li>

<li><p><a href="/articles/vendor/github.com/shurcooL/sanitized_anchor_name/README.md">vendor/github.com/shurcooL/sanitized_anchor_name/README.md</a></p></li>
</ul>

		</article>
	</body>
</html>`)

	actual = bytes.Replace(actual, []byte("%2F"), []byte("/"), -1)
	actual = bytes.Replace(actual, []byte("+"), []byte(" "), -1)

	if !bytes.Equal(actual, expect) {
		t.Errorf("Not same: `%s`", actual)
	}
}
