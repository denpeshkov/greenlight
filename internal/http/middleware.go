package http

import (
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/denpeshkov/greenlight/internal/greenlight"
	"golang.org/x/time/rate"
)

// statusResponseWriter records status of the HTTP response.
type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func newStatusResponseWriter(w http.ResponseWriter) *statusResponseWriter {
	// WriteHeader() is not called if our response implicitly returns 200 OK, so we default to that status code.
	return &statusResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// Unwrap is used by a [http.ResponseController].
func (w *statusResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

type notFoundResponseWriter struct {
	*statusResponseWriter
}

func (w *notFoundResponseWriter) Write(data []byte) (n int, err error) {
	if w.statusResponseWriter.status == http.StatusNotFound {
		data, err = json.Marshal(ErrorResponse{Msg: greenlight.ErrNotFound.Msg})
	}
	if err != nil {
		return 0, err
	}
	return w.statusResponseWriter.Write(data)
}

// notFound returns a request handler that handles [http.StatusNotFound] status code.
func (s *Server) notFound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nfw := &notFoundResponseWriter{statusResponseWriter: newStatusResponseWriter(w)}
		h.ServeHTTP(nfw, r)
	})
}

type methodNotAllowedResponseWriter struct {
	*statusResponseWriter
}

func (w *methodNotAllowedResponseWriter) Write(data []byte) (n int, err error) {
	if w.statusResponseWriter.status == http.StatusMethodNotAllowed {
		data, err = json.Marshal(ErrorResponse{Msg: http.StatusText(http.StatusMethodNotAllowed)})
	}
	if err != nil {
		return 0, err
	}
	return w.statusResponseWriter.Write(data)
}

// methodNotAllowed returns a request handler that handles [http.StatusMethodNotAllowed] status code.
func (s *Server) methodNotAllowed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mrw := &methodNotAllowedResponseWriter{statusResponseWriter: newStatusResponseWriter(w)}
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
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mw, ok := w.(*statusResponseWriter)
		if !ok {
			mw = newStatusResponseWriter(w)
		}

		start := time.Now()
		totalRequestsReceived.Add(1)

		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		totalResponsesSentByStatus.Add(strconv.Itoa(mw.status), 1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}
