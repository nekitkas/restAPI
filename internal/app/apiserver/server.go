package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/handlers"
	"github.com/gorilla/sessions"
	"github.com/nekitkas/restAPI/internal/app/model"
	"github.com/nekitkas/restAPI/internal/app/store"
	"github.com/nekitkas/restAPI/pkg/router"
	"log"
	"net/http"
	"time"
)

const (
	sessionName        = "session"
	ctxKeyUser  ctxKey = iota
	ctxKeyRequestID
)

type ctxKey int8

type server struct {
	router       *router.Router
	logger       *log.Logger
	store        store.Store
	sessionStore sessions.Store
}

func newServer(store store.Store, sessionStore sessions.Store) *server {
	s := &server{
		router:       router.NewRouter(),
		logger:       log.Default(),
		store:        store,
		sessionStore: sessionStore,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.Use(s.setRequestID)
	s.router.Use(s.logRequest)
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))

	s.router.HandleFunc("POST", "/users", s.handleUsersCreate())
	s.router.HandleFunc("GET", "/users/:id", s.handleGetUser())
	s.router.HandleFunc("POST", "/sessions", s.handleSessionsCreate())

	s.router.UseWithPrefix("/private", s.authenticateUser)
	s.router.HandleFunc("GET", "/private/profile", s.handleProfile())
}

func (s *server) setRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriter{w, http.StatusOK}
		s.logger.Printf("started %s %s\nremote_addr:%s  request_id:%s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			r.Context().Value(ctxKeyRequestID),
		)
		start := time.Now()
		next.ServeHTTP(rw, r)
		s.logger.Printf("completed in %s with %d %s\nremote_addr:%s  request_id:%s",
			time.Now().Sub(start),
			rw.code,
			http.StatusText(rw.code),
			r.RemoteAddr,
			r.Context().Value(ctxKeyRequestID),
		)
	})
}

// Middleware for auth
func (s *server) authenticateUser(next http.Handler) http.Handler {
	fmt.Println("method authenticateUser")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
		}

		id, ok := session.Values["user_id"]
		if !ok {
			fmt.Println("1")
			s.error(w, r, http.StatusUnauthorized, errors.New("not authenticated"))
			return
		}

		u, err := s.store.User().Find(id.(int))
		if err != nil {
			fmt.Println("2")
			s.error(w, r, http.StatusUnauthorized, errors.New("not authenticated"))
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyUser, u)))
	})
}

func (s *server) handleProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.respond(w, r, http.StatusOK, r.Context().Value(ctxKeyUser).(*model.User))
	}
}

func (s *server) handleUsersCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		user := &model.User{
			Email:    req.Email,
			Password: req.Password,
		}
		if err := s.store.User().Create(user); err != nil {
			s.error(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		user.Sanitize()
		s.respond(w, r, http.StatusCreated, user)
		s.logger.Printf("User %v created", user)
	}
}

func (s *server) handleGetUser() http.HandlerFunc {
	type respond struct {
		ID string `json:"id"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		id := router.Param(r.Context(), "id")
		if id == "" {
			s.error(w, r, http.StatusNoContent, errors.New("no user with such id"))
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		fmt.Println("bla bla", session.Values["user_id"])
		s.respond(w, r, http.StatusOK, respond{id})
	}
}

func (s *server) handleSessionsCreate() http.HandlerFunc {
	type request struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			s.error(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.store.User().FindByEmail(req.Email)
		if err != nil || !u.ComparePassword(req.Password) {
			s.error(w, r, http.StatusUnauthorized, errors.New("incorrect email or password"))
			return
		}

		session, err := s.sessionStore.Get(r, sessionName)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}

		session.Values["user_id"] = u.ID
		err = s.sessionStore.Save(r, w, session)
		if err != nil {
			s.error(w, r, http.StatusInternalServerError, err)
			return
		}
		fmt.Println("check if session value is saved:", session.Values["user_id"])
		s.respond(w, r, http.StatusOK, nil)
	}
}

func (s *server) error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *server) respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
