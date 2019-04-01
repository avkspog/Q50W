package main

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

type ViewData struct {
	Client Client
}

type Client struct {
	ID string
}

var settings *Config

var (
	cookieExpires = 1 * 365 * 24 * time.Hour
	tmpl          = template.Must(template.ParseFiles(path.Join("templates", "index.html")))
)

func NewHttpServer(cfg *Config) *http.Server {
	settings = cfg

	handler := createHandler()

	s := &http.Server{
		Addr:           settings.Addr(),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s
}

func createHandler() *http.ServeMux {
	handler := http.NewServeMux()

	fs := http.FileServer(http.Dir("static"))
	handler.Handle("/static/", http.StripPrefix("/static/", fs))

	handler.HandleFunc("/", handleIndex)
	handler.HandleFunc("/setwatch", handleSetCookie)

	return handler
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	var watchIDValue string
	viewData := ViewData{}

	cookie, err := r.Cookie(settings.CookieIDName)
	if err != nil {
		handleIndexTemplate(w, r, viewData)
		return
	}

	watchIDValue = spaceJoin(cookie.Value)

	if !(len(watchIDValue) > 0 && len(watchIDValue) <= settings.CookieMaxLength) {
		cookie := &http.Cookie{
			Name:   settings.CookieIDName,
			Path:   "/",
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		handleIndexTemplate(w, r, viewData)
		return
	}

	//TODO add request to map service
	client := Client{}
	client.ID = watchIDValue
	viewData.Client = client
	handleIndexTemplate(w, r, viewData)
}

func handleIndexTemplate(w http.ResponseWriter, r *http.Request, data interface{}) {
	if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func handleSetCookie(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	r.ParseForm()
	watchIDValue := spaceJoin(r.FormValue(settings.CookieIDName))

	if !(len(watchIDValue) > 0 && len(watchIDValue) <= settings.CookieMaxLength) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}

	cookie := &http.Cookie{
		Name:    settings.CookieIDName,
		Value:   watchIDValue,
		Path:    "/",
		Expires: time.Now().Add(cookieExpires),
	}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func spaceJoin(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func (c Client) IsDefined() bool {
	return c.ID != "" && len(c.ID) > 0
}
