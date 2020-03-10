package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
)

// Page data struct for page render
type Page struct {
	Title string
	Body  []byte
}

// ListPageData list of Page
type ListPageData struct {
	PageTitle string
	Todos     []Page
}

const baseTemplatePath = "template/"
const baseNotesPath = "notes/"

func (p *Page) save() error {
	filename := baseNotesPath + p.Title + ".txt"
	// user read only permission
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := baseNotesPath + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseFiles("template/edit.html", "template/view.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir("./notes")
	if err != nil {
		log.Fatal(err)
	}

	pages := []Page{}
	for _, file := range files {
		fn := file.Name()
		p := Page{Title: strings.TrimSuffix(fn, path.Ext(fn))}
		pages = append(pages, p)
	}

	data := ListPageData{
		PageTitle: "My notes list",
		Todos:     pages,
	}

	tmpl := template.Must(template.ParseFiles("template/list.html"))
	err = tmpl.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Println("add method:", r.Method) //get request method
	if r.Method == "GET" {
		tmpl := template.Must(template.ParseFiles("template/add.html"))
		err := tmpl.Execute(w, nil)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	} else {
		title := r.FormValue("title")
		body := r.FormValue("body")
		p := &Page{Title: title, Body: []byte(body)}
		err := p.save()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/view/"+title, http.StatusFound)
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/add", addHandler)
	log.Println("The application is running on port 8686...")
	log.Fatal(http.ListenAndServe(":8686", nil))
}
