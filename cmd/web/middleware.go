package main

import (
	"crypto/subtle"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy",
			"default-src 'self' https://cdn.jsdelivr.net/npm/; "+
				"style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net/npm/; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net/npm/;"+
				"font-src 'self' https://cdn.jsdelivr.net/npm/; "+
				"img-src * data: https://online.swagger.io;",
		)
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.Info("received request",
			"remote_address", r.RemoteAddr,
			"protocol", r.Proto,
			"request_method", r.Method,
			"request_url", r.URL.String(),
		)
		rID := uuid.New()
		ctx := AppendCtx(r.Context(), slog.String("request_id", rID.String()))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverError(w, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		time.Sleep(time.Minute)

		mu.Lock()

		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}

		mu.Unlock()
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverError(w, fmt.Errorf("%s", err))
			return
		}

		mu.Lock()

		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(25, 100)}
		}

		clients[ip].lastSeen = time.Now()

		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.serverError(w, fmt.Errorf("%s", err))
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func (app *application) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(app.config.Server.User)) != 1 ||
			subtle.ConstantTimeCompare([]byte(pass), []byte(app.config.Server.Password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="`+app.config.Server.Realm+`"`)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
