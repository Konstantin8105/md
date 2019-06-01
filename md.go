package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/russross/blackfriday"
)

var tmpl = `
<html>
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			.markdown-body {
				box-sizing: border-box;
				min-width: 200px;
				max-width: 980px;
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
			%s
		</article>
	</body>
</html>`

func main() {
	// create flags
	help := flag.Bool("h", false, "print help information")
	port := flag.String("p", "8080", "server port")

	// parsing flags
	flag.Parse()

	flag.Usage = func() {
		flag.PrintDefaults()
	}

	// flag action
	if *help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	// output used server port
	fmt.Fprintf(os.Stdout, "Start server on port :%s\n", *port)

	// prepare server

	// generate main page
	http.HandleFunc("/", mainHandler)
	// generate articles
	http.HandleFunc("/articles/", articleHandler)

	// static files for example: pictures
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// start server
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Server error : %v", err)
	}
}

// mainHandler generate main web page with list of articles
func mainHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "GET : %v\n", r.URL.Path)

	if err := func() (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("Try open page: %v. %v", r.URL.Path, err)
			}
		}()
		// find all markdown files
		var mdFiles []string
		if err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err != nil {
				fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			// get filename markdown files
			if strings.HasSuffix(info.Name(), ".md") {
				mdFiles = append(mdFiles, info.Name())
			}
			return nil
		}); err != nil {
			return fmt.Errorf("Cannot walk: %v", err)
		}
		// generate markdown main page
		var mainTmpl string = "# List of articles:\n"
		for _, file := range mdFiles {
			// get name of article
			var name string = file

			// add to main page
			mainTmpl += fmt.Sprintf("\n* [%s](/articles/%s)\n", name, file)
		}
		// generate html by markdown
		html := blackfriday.MarkdownCommon([]byte(mainTmpl))
		fmt.Fprintf(w, tmpl, html)
		return
	}(); err != nil {
		fmt.Fprintf(w, "Error : %v", err)
	}
}

// articleHandler generate web page with article
func articleHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "GET : %v\n", r.URL.Path)

	if err := func() (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("Try open page: %v. %v", r.URL.Path, err)
			}
		}()
		// get title
		title := r.URL.Path[len("/articles/"):]
		title = strings.TrimSpace(title)
		if title == "" {
			err = fmt.Errorf("Title of article is empty")
			return
		}
		// get file content
		var content []byte
		content, err = ioutil.ReadFile(title)
		if err != nil {
			err = fmt.Errorf("Cannot read file: %v", err)
			return
		}
		// generate markdown
		html := blackfriday.MarkdownCommon(content)
		fmt.Fprintf(w, tmpl, html)
		return
	}(); err != nil {
		fmt.Fprintf(w, "Error : %v", err)
	}
}
