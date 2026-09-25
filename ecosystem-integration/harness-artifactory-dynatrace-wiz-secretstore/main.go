package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/antonkazachenko/go-todo-list-api/config"
	"github.com/antonkazachenko/go-todo-list-api/internal/service"
	storage "github.com/antonkazachenko/go-todo-list-api/internal/storage/sqlite"
	"github.com/antonkazachenko/go-todo-list-api/routes"
	"github.com/golang-jwt/jwt/v4"
)

func isValidToken(tokenString string) bool {
	if tokenString == "" {
		return false
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.GetPassword()), nil
	})
	if err != nil || !token.Valid {
		return false
	}
	return true
}

func main() {
	db := storage.InitDB()
	defer db.Close()

	taskRepo := storage.NewSQLiteTaskRepository(db)

	taskService := service.NewTaskService(taskRepo)

	router := routes.RegisterRoutes(taskService)

	// Default route: unauthenticated users see the branded login screen.
	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || !isValidToken(cookie.Value) {
			http.Redirect(w, r, "/login.html", http.StatusFound)
			return
		}
		http.ServeFile(w, r, "./web/index.html")
	})

	router.Get("/login", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login.html", http.StatusFound)
	})

	// Health endpoint for Kubernetes readiness/liveness checks
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Static file server (fallback)
	fileServer := http.FileServer(http.Dir("./web"))
	router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if filepath.Ext(r.URL.Path) == ".css" {
			w.Header().Set("Content-Type", "text/css")
		}
		fileServer.ServeHTTP(w, r)
	})

	address := fmt.Sprintf(":%s", config.TODO_PORT)
	log.Printf("Starting server on %s", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
