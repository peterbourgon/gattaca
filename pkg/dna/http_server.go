package dna

import (
	"fmt"
	"net/http"
	"strings"
)

// HTTPServer wraps a Service and implements http.Handler.
type HTTPServer struct {
	service Service
}

// NewHTTPServer returns an HTTPServer wrapping the Service.
func NewHTTPServer(service Service) *HTTPServer {
	return &HTTPServer{
		service: service,
	}
}

// ServeHTTP implements http.Handler.
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		first  = extractPathToken(r.URL.Path, 0)
		method = r.Method
	)
	switch {
	case method == "POST" && first == "add":
		var (
			user     = r.URL.Query().Get("user")
			token    = r.URL.Query().Get("token")
			sequence = r.URL.Query().Get("sequence")
		)
		err := s.service.Add(r.Context(), user, token, sequence)
		switch {
		case err == nil:
			fmt.Fprintln(w, "Add OK")
		case err == ErrBadAuth:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	case method == "GET" && first == "check":
		var (
			user        = r.URL.Query().Get("user")
			token       = r.URL.Query().Get("token")
			subsequence = r.URL.Query().Get("subsequence")
		)
		err := s.service.Check(r.Context(), user, token, subsequence)
		switch {
		case err == nil:
			fmt.Fprintln(w, "Subsequence found")
		case err == ErrSubsequenceNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case err == ErrBadAuth:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

	default:
		http.NotFound(w, r)
	}
}

func extractPathToken(path string, position int) string {
	toks := strings.Split(strings.Trim(path, "/ "), "/")
	if len(toks) <= position {
		return ""
	}
	return toks[position]
}
