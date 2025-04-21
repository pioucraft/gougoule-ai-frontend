package main

import (
	"fmt"
	"framework/api"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var htmlAppFile []byte

func loader() {
	htmlFile, err := os.ReadFile("src/app.html")
	if err != nil {
		log.Fatalf("Failed to read src/app.html: %v", err)
	}
	htmlAppFile = htmlFile // Assign the file content to the global variable

}

func pageHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api") {
		api.API(w, r)
		return
	}

	if r.URL.Path == "/script.js" {
		http.ServeFile(w, r, "script.js")
		return
	}

	if r.URL.Path == "/default.css" {
		http.ServeFile(w, r, "default.css")
		return
	}

	if r.URL.Path == "/" {
		indexPath := filepath.Join("index.html")
		returnHTML(w, indexPath)
		return
	}
	rawPaths := strings.Split(r.URL.Path, "/")
	paths := []string{}
	for _, p := range rawPaths {
		if p != "" {
			paths = append(paths, p)
		}
	}
	currentPath := ""

	for i, segment := range paths {

		dirPath := filepath.Join("src/app", currentPath)
		dir, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}

		found := false
		for _, entry := range dir {
			if entry.Name() == segment {
				found = true
				currentPath = filepath.Join(currentPath, entry.Name())
				break
			}
		}
		if !found {
			for _, entry := range dir {
				if entry.Name() == "[slug]" {
					found = true
					currentPath = filepath.Join(currentPath, "[slug]")
					break
				}
			}
		}
		if !found {
			tryToGetStaticFile(w, r)
			return
		}

		if i == len(paths)-1 {
			path := filepath.Join(currentPath, "index.html")
			returnHTML(w, path)
			return
		}

	}

}

func main() {
	loader()

	http.HandleFunc("/", pageHandler)
	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func returnHTML(w http.ResponseWriter, path string) {
	htmlFile, err := os.ReadFile(filepath.Join("src/app/", path))
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	for strings.Contains(string(htmlFile), "{@component}") {
		index := strings.Index(string(htmlFile), "{@component}")
		endIndex := strings.Index(string(htmlFile)[index:], "{/component}")
		if endIndex == -1 {
			http.Error(w, "Component not closed", http.StatusInternalServerError)
			return
		}
		endIndex += index + len("{/component}")
		componentPath := string(htmlFile)[index+len("{@component}") : endIndex-len("{/component}")]
		component, err := os.ReadFile(filepath.Join("src/components", componentPath+".html"))
		if err != nil {
			http.Error(w, "Component not found", http.StatusNotFound)
			return
		}
		htmlFile = []byte(strings.Replace(string(htmlFile), "{@component}"+componentPath+"{/component}", string(component), 1))
	}

	html := "<script src='/script.js'></script>" + "<link rel='stylesheet' href='/default.css'></link>" +
		strings.ReplaceAll(string(htmlAppFile), "{@app}", string(htmlFile))
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func tryToGetStaticFile(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	filePath := filepath.Join("src/static", path)

	{
		http.ServeFile(w, r, filePath)
		return
	}
}
