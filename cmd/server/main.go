package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/petaki/inertia-go"

	"github.com/rluders/go-inertiajs-react/internal/config"
	"github.com/rluders/go-inertiajs-react/internal/handler"
	"github.com/rluders/go-inertiajs-react/resources/views"
	"github.com/rluders/go-inertiajs-react/static"
)

var inertiaManager *inertia.Inertia

func main() {
	// Configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Could't read config: %s", err)
		os.Exit(-1)
	}

	inertiaManager = loadInertia(cfg.Inertia)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv, err := newServer(cfg.Server)
	if err != nil {
		log.Fatal(err)
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

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}

	log.Print("Server exited properly")
}

func newServer(cfg *config.Server) (*http.Server, error) {
	router := mux.NewRouter()
	router.Use(inertiaManager.Middleware)
	// add inertia manager to context
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "inertia", inertiaManager)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	router.HandleFunc("/", handler.HomeHandler).Methods(http.MethodGet)

	// Static file server
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(static.Files))))

	return &http.Server{
		Handler:           router,
		Addr:              cfg.Host,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       1 * time.Second,
	}, nil
}

func loadInertia(cfg *config.Inertia) *inertia.Inertia {
	inertiaManager := inertia.NewWithFS(cfg.URL, cfg.RootTemplate, cfg.Version, views.Templates)

	// handle all shared values
	if len(cfg.Shared) > 0 {
		for k, v := range cfg.Shared {
			key := strings.ToLower(k)
			inertiaManager.Share(key, v)
		}
	}

	inertiaManager.ShareFunc("mix", func(path string) (string, error) {
		return "/" + path, nil
	})

	return inertiaManager
}

func loadConfig() (*config.Config, error) {
	// Configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config.yaml"
	}

	cfg, err := config.NewConfig(configPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
