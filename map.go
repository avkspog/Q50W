package main

import (
	pb "Q50W/api"
	"context"
	"html/template"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"google.golang.org/grpc"
)

type ViewData struct {
	Client Client
}

type Client struct {
	ID    string
	Point *pb.Point
}

const (
	STATIC_DIR = "/static/"
)

var settings *Config

var (
	cookieExpires = 1 * 365 * 24 * time.Hour
	tmpl          = template.Must(template.ParseFiles(path.Join("templates", "index.html")))
	checkClientID = regexp.MustCompile("^[A-Za-z0-9]{1,15}$").MatchString
)

func NewHTTPServer(cfg *Config) *http.Server {
	settings = cfg

	handler := Router()

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

func Router() *mux.Router {
	router := mux.NewRouter()

	router.PathPrefix(STATIC_DIR).Handler(http.StripPrefix(STATIC_DIR, http.FileServer(http.Dir("."+STATIC_DIR)))).Methods("GET")
	router.HandleFunc("/", handleIndex).Methods("GET")
	router.HandleFunc("/set_id", handleSetCookie).Methods("POST")

	return router
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
	result := checkClientID(watchIDValue)

	if !result {
		cookie := &http.Cookie{
			Name:   settings.CookieIDName,
			Path:   "/",
			MaxAge: -1,
		}
		http.SetCookie(w, cookie)
		handleIndexTemplate(w, r, viewData)
		return
	}

	point := grpcWatchServiceCall(watchIDValue)

	client := Client{}
	client.ID = watchIDValue
	if point != nil {
		client.Point = point
	}
	viewData.Client = client

	handleIndexTemplate(w, r, viewData)
}

func handleSetCookie(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	watchIDValue := spaceJoin(r.FormValue(settings.CookieIDName))
	result := checkClientID(watchIDValue)

	if !result {
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

func handleIndexTemplate(w http.ResponseWriter, r *http.Request, data interface{}) {
	if err := tmpl.ExecuteTemplate(w, "index", data); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}

func spaceJoin(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func (c Client) IsDefined() bool {
	return c.ID != "" && len(c.ID) > 0
}

func (c Client) HasPoint() bool {
	return c.Point != nil && (c.Point.GetLatitude() > 0 && c.Point.GetLongitude() > 0)
}

func grpcWatchServiceCall(clientID string) *pb.Point {
	conn, err := grpc.Dial(settings.ServiceAddr(), grpc.WithInsecure())
	if err != nil {
		log.Printf("%v", err)
		return nil
	}

	defer conn.Close()

	client := pb.NewRoutePointClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	point, err := client.LastPoint(ctx, &pb.Identifier{Version: "1", ClientId: clientID})
	if err != nil {
		log.Printf("%v", err)
		return nil
	}

	return point
}
