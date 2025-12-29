package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
    "net/http"
    "time"

    "github.com/google/uuid"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
)

type User struct {
	ID          []byte
	Username    string
	Credentials []webauthn.Credential
}

func (u *User) WebAuthnID() []byte {
	return u.ID
}

func (u *User) WebAuthnName() string {
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	return u.Username
}

func (u *User) WebAuthnIcon() string {
	return ""
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

type Service struct {
	db  *sql.DB
	wan *webauthn.WebAuthn
}

func NewService(db *sql.DB, host string) (*Service, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: "365 Project",
		RPID:          host, // "localhost" or domain
		RPOrigins:     []string{
            fmt.Sprintf("http://%s", host), 
            fmt.Sprintf("https://%s", host),
            fmt.Sprintf("http://%s:8080", host),
            "http://localhost:8080", 
        },
	}

	wan, err := webauthn.New(wconfig)
	if err != nil {
		return nil, err
	}

	return &Service{
		db:  db,
		wan: wan,
	}, nil
}

// --- User Management ---

func (s *Service) GetUserCount() (int, error) {
    var count int
    err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    return count, err
}

func (s *Service) GetUser(username string) (*User, error) {
    var user User
    var credentialData []byte
    err := s.db.QueryRow("SELECT id, username, credentials FROM users WHERE username = ?", username).Scan(&user.ID, &user.Username, &credentialData)
    if err != nil {
        return nil, err
    }
    // user.DisplayName is not a field, it's a method
    if len(credentialData) > 0 {
        json.Unmarshal(credentialData, &user.Credentials)
    }
    return &user, nil
}

func (s *Service) CreateUser(username string) (*User, error) {
    // Check if exists
    if _, err := s.GetUser(username); err == nil {
        return nil, fmt.Errorf("user already exists")
    }

    user := &User{
        ID:          []byte(username), // Simple ID
        Username:    username,
        Credentials: []webauthn.Credential{},
    }
    
    // Save to DB
    err := s.SaveUser(user)
    return user, err
}

func (s *Service) SaveUser(user *User) error {
    credsBlob, err := json.Marshal(user.Credentials)
    if err != nil {
        return err
    }
    _, err = s.db.Exec(`
        INSERT INTO users (id, username, credentials) VALUES (?, ?, ?)
        ON CONFLICT(username) DO UPDATE SET credentials=excluded.credentials
    `, user.ID, user.Username, credsBlob)
    return err
}

// --- Sessions ---

func (s *Service) CreateSession(userID []byte) (string, error) {
    token := uuid.New().String()
    expires := time.Now().Add(24 * 30 * time.Hour) // 30 days
    _, err := s.db.Exec("INSERT INTO sessions (token, user_id, expires_at) VALUES (?, ?, ?)", token, string(userID), expires)
    return token, err
}

func (s *Service) ValidateSession(token string) (bool, error) {
    var exists bool
    err := s.db.QueryRow("SELECT EXISTS(SELECT 1 FROM sessions WHERE token = ? AND expires_at > ?)", token, time.Now()).Scan(&exists)
    return exists, err
}

// Registration
func (s *Service) BeginRegistration(user *User) (*protocol.CredentialCreation, *webauthn.SessionData, error) {
	return s.wan.BeginRegistration(user)
}

func (s *Service) FinishRegistration(user *User, session webauthn.SessionData, r *http.Request) (*webauthn.Credential, error) {
	return s.wan.FinishRegistration(user, session, r)
}

// Login
func (s *Service) BeginLogin(user *User) (*protocol.CredentialAssertion, *webauthn.SessionData, error) {
	return s.wan.BeginLogin(user)
}

func (s *Service) FinishLogin(user *User, session webauthn.SessionData, r *http.Request) (*webauthn.Credential, error) {
	return s.wan.FinishLogin(user, session, r)
}
