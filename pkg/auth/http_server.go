package auth

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// HTTPServer wraps a Service and implements http.Handler.
type HTTPServer struct {
	router  *mux.Router
	service Service
}

// NewHTTPServer returns an HTTPServer wrapping the Service.
func NewHTTPServer(service Service) *HTTPServer {
	s := &HTTPServer{
		service: service,
	}
	r := mux.NewRouter()
	{
		r.Methods("POST").Path("/signup").HandlerFunc(s.handleSignup)
		r.Methods("POST").Path("/login").HandlerFunc(s.handleLogin)
		r.Methods("GET").Path("/validate").HandlerFunc(s.handleValidate)
		r.Methods("POST").Path("/logout").HandlerFunc(s.handleLogout)
	}
	s.router = r
	return s
}

// ServeHTTP implements http.Handler, delegating to the mux.Router.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *HTTPServer) handleSignup(w http.ResponseWriter, r *http.Request) {
	var (
		user = r.URL.Query().Get("user")
		pass = r.URL.Query().Get("pass")
	)
	if err := s.service.Signup(r.Context(), user, pass); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "signup successful")
}

func (s *HTTPServer) handleLogin(w http.ResponseWriter, r *http.Request) {
	var (
		user = r.URL.Query().Get("user")
		pass = r.URL.Query().Get("pass")
	)
	token, err := s.service.Login(r.Context(), user, pass)
	if err == ErrBadAuth {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, token)
}

func (s *HTTPServer) handleValidate(w http.ResponseWriter, r *http.Request) {
	var (
		user  = r.URL.Query().Get("user")
		token = r.URL.Query().Get("token")
	)
	err := s.service.Validate(r.Context(), user, token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	fmt.Fprintln(w, "validate successful")
}

func (s *HTTPServer) handleLogout(w http.ResponseWriter, r *http.Request) {
	var (
		user  = r.URL.Query().Get("user")
		token = r.URL.Query().Get("token")
	)
	err := s.service.Logout(r.Context(), user, token)
	if err == ErrBadAuth {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "logout successful")
}
