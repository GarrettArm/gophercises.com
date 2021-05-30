package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
)

type Story map[string]Page
type Page struct {
	Title   string   `json:"title"`
	Story   []string `json:"story"`
	Options []struct {
		Text string `json:"text"`
		Arc  string `json:"arc"`
	} `json:"options"`
}

func ParseJSON(filepath string) (Story, error) {
	var s Story
	b, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, &s)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func parseFlags() (filename *string, port *int) {
	filename = flag.String("f", "gopher.json", "a json file containing a story")
	port = flag.Int("p", 3030, "port on which cyoa is served")
	flag.Parse()
	return filename, port
}

func storyHandler(s Story, t *template.Template) http.Handler {
	if t == nil {
		t = template.Must(template.ParseFiles("story.tmpl"))
	}
	return handler{s, t}
}

type handler struct {
	s Story
	t *template.Template
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "" || path == "/" {
		path = "/intro"
	}
	path = path[1:]
	if page, ok := h.s[path]; ok {
		t := template.Must(template.ParseFiles("story.tmpl"))
		err := t.Execute(w, page)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Server error", http.StatusInternalServerError)
		}
		return
	}
	http.Error(w, "Page Not Found", http.StatusNotFound)
}

func main() {
	filename, port := parseFlags()

	story, err := ParseJSON(*filename)
	if err != nil {
		fmt.Println("Error processing source json")
		panic(err)
	}

	fmt.Printf("Serving on %d\n", *port)
	template := template.Must(template.ParseFiles("story.tmpl"))
	handler := storyHandler(story, template)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), handler))
}
