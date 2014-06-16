// @purplemaze
// GO wiki!
package main

import (
	"html/template"
	"io/ioutil"
	"net/http"
	"regexp"
)

// The Page struct describes how page data will be stored in memory.
type Page struct {
	Title string
	Body  []byte
}

//Global vars
var templates = template.Must(template.ParseFiles("tmpl/edit.html", "tmpl/view.html"))

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

// This method will save the Page's Body to a text file. For simplicity, we will use the Title as the file name.
func (p *Page) save() error {
	filename := "data/" + p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

// The function loadPage constructs the file name from the title parameter, reads the file's contents into a new variable body,
// and returns a pointer to a Page literal constructed with the proper title and body values.
func loadPage(title string) (*Page, error) {
	filename := "data/" + title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// First, this function extracts the page title from r.URL.Path, the path component of the request URL.
// The Path is re-sliced with [len("/view/"):] to drop the leading "/view/" component of the request path.
// This is because the path will invariably begin with "/view/", which is not part of the page's title.
// The function then loads the page data, formats the page with a string of simple HTML,
// and writes it to w, the http.ResponseWriter.
func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound) // The http.Redirect function adds an HTTP status code of http.StatusFound (302) and a Location header to the HTTP response.
		return
	}
	renderTemplate(w, "view", p)
}

// The function template.ParseFiles will read the contents of edit.html and return a *template.Template.
// The method t.Execute executes the template, writing the generated HTML to the http.ResponseWriter.
// The .Title and .Body dotted identifiers refer to p.Title and p.Body.
func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

// The function saveHandler will handle the submission of forms located on the edit pages.
// After uncommenting the related line in main, let's implement the the handler:
// The page title (provided in the URL) and the form's only field, Body, are stored in a new Page.
// The save() method is then called to write the data to a file, and the client is redirected to the /view/ page.
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

// A wrapper function that takes a function of the above type, and returns a function of type
// http.HandlerFunc (suitable to be passed to the function http.HandleFunc)
// The returned function is called a closure because it encloses values defined outside of it.
// In this case, the variable fn (the single argument to makeHandler) is enclosed by the closure.
// The variable fn will be one of our save, edit, or view handlers.
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

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.ListenAndServe(":8080", nil)
}
