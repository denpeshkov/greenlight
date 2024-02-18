package http

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"golang.org/x/time/rate"
)

// hijackResponseWriter records status of the HTTP response.
type hijackResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *hijackResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

type notFoundResponseWriter struct {
	hijackResponseWriter
}

func (w *notFoundResponseWriter) Write(data []byte) (n int, err error) {
	if w.hijackResponseWriter.status == http.StatusNotFound {
		data, err = json.Marshal(ErrorResponse{Msg: greenlight.ErrNotFound.Msg})
	}
	if err != nil {
		return 0, err
	}
	return w.hijackResponseWriter.Write(data)
}

// notFound returns a request handler that handles [http.StatusNotFound] status code.
func (s *Server) notFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
		nfw := &notFoundResponseWriter{hijackResponseWriter: hw}
		h.ServeHTTP(nfw, r)
	})
}

type methodNotAllowedResponseWriter struct {
	hijackResponseWriter
}

func (w *methodNotAllowedResponseWriter) Write(data []byte) (n int, err error) {
	if w.hijackResponseWriter.status == http.StatusMethodNotAllowed {
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusMethodNotAllowed)})
	}
	if err != nil {
		return 0, err
	}
	return w.hijackResponseWriter.Write(data)
}

// methodNotAllowed returns a request handler that handles [http.StatusMethodNotAllowed] status code.
func (s *Server) methodNotAllowed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hw := hijackResponseWriter{ResponseWriter: w, status: http.StatusOK}
		mrw := &methodNotAllowedResponseWriter{hijackResponseWriter: hw}
		h.ServeHTTP(mrw, r)
	})
}

func (s *Server) recoverPanic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Acts as a trigger to make HTTP server automatically close the current connection after a response has been sent.
				w.Header().Set("Connection", "close")
				s.Error(w, r, fmt.Errorf("%v", err))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func (s *Server) rateLimit(h http.Handler) http.Handler {
	op := "http.Server.rateLimit"

	type clientLim struct {
		lim      *rate.Limiter
		lastSeen time.Time
	}

	lims := sync.Map{}

	// cleanup unused limiters
	go func() {
		for {
			time.Sleep(time.Minute)

			lims.Range(func(ip, v any) bool {
				clim := v.(clientLim)
				if time.Since(clim.lastSeen) > 3*time.Minute {
					lims.Delete(ip)
				}
				return true
			})
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			s.Error(w, r, fmt.Errorf("%s: %w", op, err))
			return
		}

		v, _ := lims.LoadOrStore(ip, clientLim{lim: rate.NewLimiter(rate.Limit(s.opts.limiterRps), s.opts.limiterBurst)})
		clim := v.(clientLim)

		clim.lastSeen = time.Now()
		if !clim.lim.Allow() {
			s.Error(w, r, greenlight.NewRateLimitError("Rate limit exceeded."))
			s.Log(w, r, "Rate limit exceeded.")
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (s *Server) authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		op := "http.Server.authenticate"

		// This indicates to any caches that the response may vary based on the value of the Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		authzHeader := r.Header.Get("Authorization")

		if authzHeader == "" {
			s.Error(w, r, fmt.Errorf("%s: %w", op, greenlight.NewUnauthorizedError("You must be authenticated to access this resource.")))
			return
		}

		headerParts := strings.Split(authzHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			s.Error(w, r, fmt.Errorf("%s: %w", op, greenlight.NewUnauthorizedError("Invalid or missing authentication token.")))
			return
		}

		userId, err := s.authService.ParseToken(headerParts[1])
		if err != nil {
			s.Error(w, r, fmt.Errorf("%s: %w", op, err))
			return
		}
		r = r.WithContext(greenlight.NewContextWithUserID(r.Context(), userId))
		h.ServeHTTP(w, r)
	})
}

func (s *Server) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
	)
	// The following code will be run for every request...
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // Record the time that we started to process the request.
		start := time.Now()
		// Use the Add() method to increment the number of requests received by 1.
		totalRequestsReceived.Add(1)
		// Call the next handler in the chain.
		next.ServeHTTP(w, r)
		// On the way back up the middleware chain, increment the number of responses // sent by 1.
		totalResponsesSent.Add(1)
		// Calculate the number of microseconds since we began to process the request, // then increment the total processing time by this amount.
		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
