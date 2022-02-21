package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/petaki/inertia-go"

	"github.com/rluders/go-inertiajs-react/resources/views"
	"github.com/rluders/go-inertiajs-react/static"
)

var inertiaManager *inertia.Inertia

func main() {
	url := "http://localhost"
	version := ""

	inertiaManager = inertia.NewWithFS(url, "app.gohtml", version, views.Templates)
	inertiaManager.Share("title", "Inertia Go with React and Laravel Mix")
	inertiaManager.ShareFunc("mix", func(path string) (string, error) {
		return "/" + path, nil
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Server started. Listen to: http://localhost:8080")

	go func() {
		err := srv.ListenAndServe()

		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-done
	log.Print("Server stopped")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("Server exited properly")
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", handle(homeHandler))

	// Static file server
	staticFS := http.FS(static.Files)
	fs := http.FileServer(staticFS)
	mux.Handle("/images/", fs)
	mux.Handle("/js/", fs)
	mux.Handle("/css/", fs)

	return mux
}

func handle(h http.HandlerFunc) http.Handler {
	return inertiaManager.Middleware(http.HandlerFunc(h))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := inertiaManager.Render(w, r, "Welcome", nil)
	if err != nil {
		log.Panic(err)
	}
}
