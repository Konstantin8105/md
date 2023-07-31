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
	"sort"
	"strings"

	"github.com/russross/blackfriday"
)

// Windows OS specific variable name
const windowsOs string = "windows"

// photos folders
const photos string = "photos"

func IsPhotos(str string) bool {
	return strings.Contains(str, string(filepath.Separator)+photos)
}

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
			img{
				max-height:500px;
				max-width:500px;
				height:auto;
				width:auto;
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
	var (
		help  = flag.Bool("h", false, "print help information")
		port  = flag.String("p", "8080", "server port")
		chdir = flag.String("ch", ".", "changes the current working directory to the named directory")
	)

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

	if err := os.Chdir(*chdir); err != nil {
		fmt.Fprintf(os.Stderr, "cannot change directory : %v", err)
	}

	// output used server port
	fmt.Fprintf(os.Stdout, "Start server on port :%s\n", *port)

	// prepare server

	// generate main page
	http.HandleFunc("/", mainHandler)
	// generate articles
	http.HandleFunc("/articles/", articleHandler)
	// generate photos
	http.HandleFunc("/"+photos+"/", photosHandler)

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
		// get all folders
		var getFolder func(string) ([]string, error)
		getFolder = func(baseFolder string) (fs []string, err error) {
			files, err := ioutil.ReadDir(baseFolder)
			if err != nil {
				return nil, err
			}
			for _, file := range files {
				if !file.IsDir() {
					continue
				}
				if file.Name() == ".git" {
					continue
				}
				fs = append(fs, baseFolder+string(os.PathSeparator)+file.Name())
			}
			size := len(fs)
			for i := 0; i < size; i++ {
				fss, err := getFolder(fs[i])
				if err != nil {
					return nil, err
				}
				fs = append(fs, fss...)
			}

			return fs, nil
		}
		folders, err := getFolder(".")
		if err != nil {
			return err
		}
		folders = append(folders, ".")
		sort.Strings(folders)

		// find all markdown files
		for i := range folders {
			files, err := ioutil.ReadDir(folders[i])
			if err != nil {
				return err
			}
			sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
			var folderHeader bool
			for _, file := range files {
				if file.IsDir() {
					continue
				}
				if !strings.HasSuffix(file.Name(), ".md") {
					continue
				}

				if !folderHeader {
					folderHeader = true
					mainTmpl += "------\n\n"
					count := strings.Count(folders[i], "\\")
					count += strings.Count(folders[i], "/")
					count++
					for i := 0; i < count && i < 3; i++ {
						mainTmpl += "#"
					}
					f := folders[i]
					if runtime.GOOS == windowsOs {
						f = strings.Replace(f, "\\", "/", -1)
					}
					mainTmpl += fmt.Sprintf(" %s\n\n", f)
				}
				path := folders[i] + string(os.PathSeparator) + file.Name()

				// Windows specific
				if runtime.GOOS == windowsOs {
					path = strings.Replace(path, "\\", "/", -1)
				}

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
								name = strings.ReplaceAll(name, "#", " ")
								name = strings.TrimSpace(name)
							}
						}
					}
				}

				// escape space
				path = url.QueryEscape(path)

				// add to main page
				mainTmpl += fmt.Sprintf("[%s](/articles/%s)\n\n", name, path)
				mainTmpl += "\n\n"
			}
		}
		mainTmpl += "------\n\n"

		// photos
		func() {
			files, err := ioutil.ReadDir(photos)
			if err != nil {
				return
			}
			mainTmpl += fmt.Sprintf("# PHOTOS\n\n")
			for _, file := range files {
				name := file.Name()
				mainTmpl += fmt.Sprintf("[%s](/photos/%s)\n\n", name, name)
				mainTmpl += "\n\n"
			}
			mainTmpl += "------\n\n"
		}()

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
				err = fmt.Errorf("Try open page in article: %v. %v", r.URL.Path, err)
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
		if runtime.GOOS == windowsOs {
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
				if runtime.GOOS == windowsOs {
					title = strings.Replace(title, "\\", "/", -1)
				}
				err = fmt.Errorf("Cannot read file `%s`: %v", title, err)
				return
			}
			//
			str := string(content)
			str = strings.Replace(str, "\r", "", -1)

			// add link to main page
			str = "[Main page](/)\n\n" + str

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

// photosHandler generate web page with photos
func photosHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(os.Stdout, "GET : %v\n", r.URL.Path)

	if err := func() (err error) {
		defer func() {
			if err != nil {
				err = fmt.Errorf("Try open page in photos: %v. %v", r.URL.Path, err)
			}
		}()
		// get title
		path := r.URL.Path
		if len(path) <= len("/"+photos+"/") {
			err = fmt.Errorf("URL path is too small: %s", path)
			return
		}
		title := path[len("/"+photos+"/")-1:]
		title = strings.TrimSpace(title)
		if title == "" {
			err = fmt.Errorf("Title of photos is empty")
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
		if runtime.GOOS == windowsOs {
			title = strings.Replace(title, "/", "\\", -1)
		}

		// fix first letter
		if len(title) > 0 && (title[0] == '\\' || title[0] == '/') {
			title = title[1:]
		}

		index := strings.LastIndex(title, string(filepath.Separator))
		if index < 0 {
			// folder list
			f := photos + string(filepath.Separator) + title
			files, err := ioutil.ReadDir(f)
			if err != nil {
				err = fmt.Errorf("readdir :`%s`. %v", f, err)
				return err
			}
			var content string
			content += "[Main page](/)\n\n"
			content += fmt.Sprintf("%s\n\n", title)
			for _, file := range files {
				content += fmt.Sprintf("![%s](/%s/%s/%s)\n\n",
					file.Name(),
					photos,
					title,
					file.Name(),
				)
			}
			html := blackfriday.Run([]byte(content))
			fmt.Fprintf(w, tmpl, html)
		} else {
			// view file
			http.ServeFile(w, r, photos+string(filepath.Separator)+title)
		}
		return
	}(); err != nil {
		fmt.Fprintf(w, "Error : %v\n", err)
	}
}
