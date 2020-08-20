package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/packago/config"
	"github.com/tullo/note.delivery/note"
	"github.com/tullo/note.delivery/templates"
	"github.com/ulule/limiter/v3"
	"github.com/ulule/limiter/v3/drivers/middleware/stdlib"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

func main() {
	log := log.New(os.Stdout, "NOTE.DELIVERY : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	r := chi.NewRouter()
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	rate, err := limiter.NewRateFromFormatted("500-H")
	if err != nil {
		panic(err)
	}
	store := memory.NewStore()
	mw := stdlib.NewMiddleware(limiter.New(store, rate, limiter.WithTrustForwardHeader(true)))
	mw.OnLimitReached = rateLimitHandler(log)

	note := note.New(log)

	r.Group(func(r chi.Router) {
		r.Use(mw.Handler)
		r.Get("/", index(log))
		r.Get("/protect-your-privacy", privacy(log))
		r.Get("/privacy-policy", policy(log))
		r.Post("/", note.Create)
		r.Get("/{noteid}", note.One)
		r.Post("/{noteid}", note.Unlock)
		r.Post("/{noteid}/delete", note.Delete)
		r.NotFound(notFound(log))
	})

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	fileServer(r, "/static", http.Dir(filepath.Join(wd, "static")))

	switch config.File().GetString("environment") {
	case "development":
		panic(http.ListenAndServe(config.File().GetString("port.development"), r))
	case "production":
		panic(http.ListenAndServe(config.File().GetString("port.production"), r))
	default:
		panic("Environment not set")
	}
}

func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}
	fs := http.StripPrefix(path, http.FileServer(root))
	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"
	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}

func notFound(log *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := templates.ReadCommonData(log, w, r)
		cd.MetaTitle = "404"
		templates.Render(w, "not-found.html", map[string]interface{}{
			"Common": cd,
		})
	}

}

func index(log *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := templates.ReadCommonData(log, w, r)
		templates.Render(w, "index.html", map[string]interface{}{
			"Common": cd,
		})
	}
}

func privacy(log *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := templates.ReadCommonData(log, w, r)
		templates.Render(w, "protect-your-privacy.html", map[string]interface{}{
			"Common": cd,
		})

	}
}

func policy(log *log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := templates.ReadCommonData(log, w, r)
		templates.Render(w, "privacy-policy.html", map[string]interface{}{
			"Common": cd,
		})
	}
}

func rateLimitHandler(log *log.Logger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cd := templates.ReadCommonData(log, w, r)
		cd.MetaTitle = "Rate Limited"
		templates.Render(w, "rate-limit.html", map[string]interface{}{
			"Common": cd,
		})
	}
}
