package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		s := <-quit

		app.infoLog.Printf("caught signal: %s", s.String())

		os.Exit(0)
	}()

	app.infoLog.Printf("staring server: %s", srv.Addr)

	return srv.ListenAndServe()
}
