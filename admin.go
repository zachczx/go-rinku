package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"rinku/templates"

	gorillaSessions "github.com/gorilla/sessions"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks"
	emailML "github.com/stytchauth/stytch-go/v16/stytch/consumer/magiclinks/email"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/sessions"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/stytchapi"
	"github.com/stytchauth/stytch-go/v16/stytch/consumer/users"
)

type AuthService struct {
	client *stytchapi.API
	store  *gorillaSessions.CookieStore
}

type User struct {
	Username string
	LoggedIn bool
}

func NewAuthService(projectId, secret string) *AuthService {
	client, err := stytchapi.NewClient(projectId, secret)
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}

	return &AuthService{
		client: client,
		store:  gorillaSessions.NewCookieStore([]byte(os.Getenv("GORILLA_SECRET_KEY"))),
	}
}

func (s *AuthService) RequireAuthentication(u *User, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.getAuthenticatedUser(w, r)
		if user != nil {
			u.Username = user.Emails[0].Email
			fmt.Println(u.Username)
		}
		if user == nil {
			http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func (s *AuthService) CheckAuthentication(u *User, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := s.getAuthenticatedUser(w, r)
		if user != nil {
			u.Username = user.Emails[0].Email
		}
		h.ServeHTTP(w, r)
	})
}

func (s *AuthService) getAuthenticatedUser(w http.ResponseWriter, r *http.Request) *users.User {
	session, err := s.store.Get(r, "stytch_session")
	if err != nil || session == nil {
		return nil
	}

	token, ok := session.Values["token"].(string)
	if !ok || token == "" {
		return nil
	}

	resp, err := s.client.Sessions.Authenticate(
		ctx,
		&sessions.AuthenticateParams{
			SessionToken: token,
		})
	if err != nil {
		delete(session.Values, "token")
		session.Save(r, w)
		return nil
	}
	session.Values["token"] = resp.SessionToken
	session.Save(r, w)

	return &resp.User
}

func (s *AuthService) sendMagicLinkHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.Form.Get("email")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	_, err := s.client.MagicLinks.Email.LoginOrCreate(
		ctx,
		&emailML.LoginOrCreateParams{
			Email: email,
		})
	if err != nil {
		log.Printf("Error sending email: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	TemplRender(w, r, templates.EmailSent())
}

func (s *AuthService) authenticateHandler(w http.ResponseWriter, r *http.Request) {
	tokenType := r.URL.Query().Get("stytch_token_type")
	token := r.URL.Query().Get("token")

	if tokenType != "magic_links" {
		log.Printf("Error: unrecognized token type %s\n", tokenType)
		http.Error(w, fmt.Sprintf("Unrecognized token type %s", tokenType), http.StatusBadRequest)
		return
	}

	resp, err := s.client.MagicLinks.Authenticate(ctx, &magiclinks.AuthenticateParams{
		Token:                  token,
		SessionDurationMinutes: 43800,
	})
	if err != nil {
		log.Printf("Error authenticating: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := s.store.Get(r, "stytch_session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["token"] = resp.SessionToken
	session.Save(r, w)

	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (s *AuthService) logout(u *User) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess, err := s.store.Get(r, "stytch_session")
		if err != nil {
			log.Printf("error getting gorilla session: %s\n", err)
		}

		sess.Options.MaxAge = -1
		delete(sess.Values, "token")
		sess.Save(r, w)

		u.Username = ""
		TemplRender(w, r, templates.LoggedOut())
	})
}
