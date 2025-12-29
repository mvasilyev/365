package main

import (
	"database/sql"
	"log"
    "io"
	"net/http"
	"os"
    "path/filepath"

    "m365/internal/api"
    "m365/internal/auth"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/mattn/go-sqlite3"
    "strings"
    "github.com/joho/godotenv"
)

func main() {
    // Load .env
    _ = godotenv.Load()

	dbPath := "photos.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

    // Initialize schema
    schema, err := os.ReadFile("internal/store/schema.sql")
    if err == nil {
        if _, err := db.Exec(string(schema)); err != nil {
            log.Printf("Error applying schema: %v", err)
        }
    } else {
        log.Printf("Warning: Could not read schema.sql: %v", err)
    }

	r := chi.NewRouter()
	r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

	// Config from Env
    appDomain := os.Getenv("APP_DOMAIN")
    if appDomain == "" { appDomain = "localhost" }

    appOrigin := os.Getenv("APP_ORIGIN")
    if appOrigin == "" { appOrigin = "http://localhost:8080" }

    // Initialize Auth
    authService, err := auth.NewService(db, appDomain, appOrigin)
    if err != nil {
        log.Fatal(err)
    }

    h := api.NewHandler(db, authService)
    h.RegisterRoutes(r)

    // Serve uploads explicitly
    workDir, _ := os.Getwd()
    uploadsDir := http.Dir(filepath.Join(workDir, "uploads"))
    FileServer(r, "/uploads", uploadsDir)

    // Serve frontend files (placeholder for now)
    filesDir := http.Dir(filepath.Join(workDir, "client/dist"))
    FileServer(r, "/", filesDir)

	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
        log.Fatal(err)
    }
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		
        // Check if file exists, otherwise serve index.html
        // This is a naive check; for better performance in production, 
        // we might want to check the specific path or let http.FileServer handle it and 404,
        // but for SPA we want 404s to go to index.html.
        
        // Simpler approach for this specific setup:
        // Attempt to serve. If 404, serve index.html manually.
        // But http.FileServer doesn't easily expose the 404 (it writes it).
        
        // Let's check using os.Stat-like logic on the directory.
        // Since 'root' is http.Dir, we can Open it.
        
        upath := r.URL.Path
        if !strings.HasPrefix(upath, pathPrefix) {
            upath = pathPrefix // simpler fallback
        }
        upath = strings.TrimPrefix(upath, pathPrefix)
        if len(upath) > 0 && upath[0] == '/' {
            upath = upath[1:]
        }
        
        f, err := root.Open(upath)
        if err != nil && os.IsNotExist(err) {
            // Serve index.html
            index, err := root.Open("index.html")
            if err == nil {
                defer index.Close()
                io.Copy(w, index) // Naive serving, good enough for demo
                return
            }
        }
        if err == nil {
            defer f.Close()
        }

		fs.ServeHTTP(w, r)
	})
}

