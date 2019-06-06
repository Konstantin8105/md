package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
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
				max-width: 900px;
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
		// generate markdown main page
		var mainTmpl string = "# List of articles:\n\n"
		// find all markdown files
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
				// Windows specific
				if runtime.GOOS == "windows" {
					path = strings.Replace(path, "\\", "/", -1)
					path = strings.Replace(path, "\\", "/", -1)
				}

				base := path

				// get name of article
				name := path
				{
					var content []byte
					content, err = ioutil.ReadFile(path)
					if err == nil {
						title := string(content[:200])
						index := strings.Index(title, "\n")
						if index > 0 {
							if title = strings.TrimSpace(title[:index]); title != "" {
								name = title
							}
						}
					}
				}

				// escape space
				path = url.QueryEscape(path)

				// add to main page
				mainTmpl += "------\n\n"
				mainTmpl += fmt.Sprintf("**Name**: %s\n\n", name)
				mainTmpl += fmt.Sprintf("**Link**: [%s](/articles/%s)\n\n",
					base, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("Cannot walk: %v", err)
		}
		// generate html by markdown
		html := blackfriday.Run([]byte(mainTmpl))
		fmt.Fprintf(w, tmpl, html)
		return
	}(); err != nil {
		fmt.Fprintf(w, "Error : %v\n", err)
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
		path := r.URL.Path
		if len(path) <= len("/articles/") {
			err = fmt.Errorf("URL path is too small: %s", path)
			return
		}
		title := path[len("/articles/")-1:]
		title = strings.TrimSpace(title)
		if title == "" {
			err = fmt.Errorf("Title of article is empty")
			return
		}
		// Unescape url
		title, err = url.QueryUnescape(title)
		if title == "" {
			err = fmt.Errorf("Cannot unescape : %v", err)
			return
		}

		// secury fix of title
		// avoid word ".."
		title = strings.ReplaceAll(title, "..", "doubledot")

		// Windows specific
		if runtime.GOOS == "windows" {
			title = strings.Replace(title, "/", "\\", -1)
		}

		// fix first letter
		if len(title) > 0 && (title[0] == '\\' || title[0] == '/') {
			title = title[1:]
		}

		// get file content
		if strings.HasSuffix(title, ".md") {
			var content []byte
			content, err = ioutil.ReadFile(title)
			if err != nil {
				err = fmt.Errorf("Cannot read file `%s`: %v", title, err)
				return
			}
			//
			str := string(content)
			str = strings.Replace(str, "\r", "", -1)

			// generate markdown
			html := blackfriday.Run([]byte(str))
			fmt.Fprintf(w, tmpl, html)
		} else {
			http.ServeFile(w, r, title)
		}
		return
	}(); err != nil {
		fmt.Fprintf(w, "Error : %v\n", err)
	}
}
