package api

import (
    "database/sql"
	"encoding/json"
    "fmt"
	"net/http"
    "io"
    "os"
    "path/filepath"
    "time"

    "m365/internal/auth"
    "m365/internal/store"

    "github.com/google/uuid"
	"github.com/go-chi/chi/v5"
    "github.com/go-webauthn/webauthn/webauthn"
    goexif "github.com/rwcarlsen/goexif/exif"
    "github.com/rwcarlsen/goexif/tiff"
    "github.com/disintegration/imaging"
)

type Handler struct {
	DB      *sql.DB
    Auth    *auth.Service
    Photos  *store.PhotoStore
    // Simple session store: username -> session data
    Sessions map[string]webauthn.SessionData 
}

func NewHandler(db *sql.DB, auth *auth.Service) *Handler {
	return &Handler{
        DB:      db,
        Auth:    auth,
        Photos:  store.NewPhotoStore(db),
        Sessions: make(map[string]webauthn.SessionData),
    }
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api", func(r chi.Router) {
		r.Get("/photos", h.ListPhotos)
        r.Group(func(r chi.Router) {
            r.Use(h.RequireAuth)
            r.Post("/photos", h.UploadPhoto)
            r.Get("/auth/status", func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`{"status":"authenticated"}`))
            })
        })
        
        // Auth routes
        r.Post("/auth/register/begin/{username}", h.BeginRegistration)
        r.Post("/auth/register/finish/{username}", h.FinishRegistration)
        r.Post("/auth/login/begin/{username}", h.BeginLogin)
        r.Post("/auth/login/finish/{username}", h.FinishLogin)
	})
}

// --- Photos ---

func (h *Handler) ListPhotos(w http.ResponseWriter, r *http.Request) {
    photos, err := h.Photos.List(365)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(photos)
}

func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
    // 10MB limit
    r.ParseMultipartForm(10 << 20) 

    file, handler, err := r.FormFile("photo")
    if err != nil {
        http.Error(w, "Error retrieving file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    day := r.FormValue("day")
    if day == "" {
        day = time.Now().Format("2006-01-02")
    }

    // Save file
    id := uuid.New().String()
    filename := fmt.Sprintf("%s%s", id, filepath.Ext(handler.Filename))
    outPath := filepath.Join("uploads", filename)
    
    // Ensure upload dir exists
    os.MkdirAll("uploads", 0755)

    dst, err := os.Create(outPath)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    if _, err := io.Copy(dst, file); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Extract EXIF
    file.Seek(0, 0)
    var lat, lon float64
    var exifJson []byte
    var orientation int
    
    x, err := goexif.Decode(file)
    if err == nil {
        // Try get coords
        if latVal, lonVal, err := x.LatLong(); err == nil {
            lat = latVal
            lon = lonVal
        }
        
        // Get orientation
        if o, err := x.Get(goexif.Orientation); err == nil {
             if val, err := o.Int(0); err == nil {
                 orientation = val
             }
        }
        fmt.Printf("Photo %s: Orientation=%d\n", id, orientation)
        
        // Serialize relevant tags (simplified)
        w := &exifWalker{exifMap: make(map[string]string)}
        x.Walk(w)
        exifJson, _ = json.Marshal(w.exifMap)
    }

    // Generate Thumbnail
    var thumbnailPath string
    // Re-open fresh for imaging to handle decoding logic safely
    thumbImg, err := imaging.Open(outPath)
    if err == nil {
        // Apply rotation based on EXIF
        // 1: Normal
        // 3: 180 rotate
        // 6: 270 rotate (Rotate 90 CW -> Rotate270 CCW)
        // 8: 90 rotate (Rotate 90 CCW -> Rotate90 CCW)
        
        switch orientation {
        case 3:
            thumbImg = imaging.Rotate180(thumbImg)
        case 6:
            // Orientation 6 = Camera rotated 90 CCW (Portrait). Needs 90 CW.
            // imaging.Rotate270 is 270 CCW = 90 CW.
            thumbImg = imaging.Rotate270(thumbImg)
        case 8:
            // Orientation 8 = Camera rotated 90 CW. Needs 90 CCW.
            thumbImg = imaging.Rotate90(thumbImg)
        }

        thumb := imaging.Fill(thumbImg, 400, 400, imaging.Center, imaging.Lanczos)
        thumbName := fmt.Sprintf("%s_thumb.jpg", id)
        thumbOutPath := filepath.Join("uploads", thumbName)
        if err := imaging.Save(thumb, thumbOutPath); err == nil {
             thumbnailPath = "/" + thumbOutPath
        }
    }

    // TODO: Resize original if too large? For now, keep original.

    p := &store.Photo{
        Day: day,
        ID: id,
        Filepath: "/" + outPath,
        ThumbnailPath: thumbnailPath,
        Notes: r.FormValue("notes"),
        Lat: lat,
        Lon: lon,
        ExifData: string(exifJson),
        CreatedAt: time.Now(),
    }

    if err := h.Photos.Save(p); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Write([]byte("Photo uploaded"))
}

type exifWalker struct {
    exifMap map[string]string
}

func (w *exifWalker) Walk(name goexif.FieldName, tag *tiff.Tag) error {
    w.exifMap[string(name)] = tag.String()
	return nil
}

// --- Auth ---

func (h *Handler) BeginRegistration(w http.ResponseWriter, r *http.Request) {
    // 1. Check if ANY user exists. If so, only allow if authenticated (or just block for now)
    count, err := h.Auth.GetUserCount()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Strict rule: Registration only allowed if 0 users.
    if count > 0 {
        http.Error(w, "Registration is closed.", http.StatusForbidden)
        return
    }

    username := chi.URLParam(r, "username")
    user, err := h.Auth.CreateUser(username)
    if err != nil {
        // If user already exists (shouldn't happen given count check, but for safety)
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    options, session, err := h.Auth.BeginRegistration(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    h.Sessions[username] = *session

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(options)
}

func (h *Handler) FinishRegistration(w http.ResponseWriter, r *http.Request) {
    username := chi.URLParam(r, "username")
    user, err := h.Auth.GetUser(username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    session, ok := h.Sessions[username]
    if !ok {
        http.Error(w, "session not found", http.StatusBadRequest)
        return
    }

    credential, err := h.Auth.FinishRegistration(user, session, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    user.Credentials = append(user.Credentials, *credential)
    if err := h.Auth.SaveUser(user); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    delete(h.Sessions, username)
    w.Write([]byte("Registration Success"))
}

func (h *Handler) BeginLogin(w http.ResponseWriter, r *http.Request) {
    username := chi.URLParam(r, "username")
    user, err := h.Auth.GetUser(username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    options, session, err := h.Auth.BeginLogin(user)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    h.Sessions[username] = *session

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(options)
}

func (h *Handler) FinishLogin(w http.ResponseWriter, r *http.Request) {
    username := chi.URLParam(r, "username")
    user, err := h.Auth.GetUser(username)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    session, ok := h.Sessions[username]
    if !ok {
        http.Error(w, "session not found", http.StatusBadRequest)
        return
    }

    credential, err := h.Auth.FinishLogin(user, session, r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Create persistent session
    token, err := h.Auth.CreateSession(user.ID)
    if err != nil {
        http.Error(w, "Failed to create session", http.StatusInternalServerError)
        return
    }

    http.SetCookie(w, &http.Cookie{
        Name:     "session_token",
        Value:    token,
        Path:     "/",
        HttpOnly: true,
        Secure:   false, // Set true in production with HTTPS
        SameSite: http.SameSiteStrictMode,
        MaxAge:   3600 * 24 * 30, // 30 days
    })

    fmt.Printf("User %s logged in with credential %v\n", username, credential.ID)

    delete(h.Sessions, username)
    w.Write([]byte("Login Success"))
}

// Middleware
func (h *Handler) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        c, err := r.Cookie("session_token")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        valid, err := h.Auth.ValidateSession(c.Value)
        if err != nil || !valid {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        next.ServeHTTP(w, r)
    })
}
