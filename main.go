package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
)

var count int64
var sema = make(chan struct{}, 1)

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	mux := http.DefaultServeMux

	mux.HandleFunc("/endpoint1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received connection to endpoint1 from", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)

		sema <- struct{}{}
		count++
		<-sema
		fmt.Println("count", count)
	})

	mux.HandleFunc("/endpoint2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received connection to endpoint2 from", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)

		sema <- struct{}{}
		count++
		<-sema
		fmt.Println("count", count)
	})

	fmt.Println("starting http listener")

	srv := http.Server{Addr: ":" + port, Handler: mux}
	go func() {
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatalln("failed to listen")
		}
	}()

	//Wait for interrupt signal to gracefully shutdown the server with a timeout of 2 seconds
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
	err := srv.Shutdown(ctx)
	if err != nil {
		logrus.WithError(err).Fatalln("graceful shutdown failed")
	}

	logrus.Infoln("closed listener")
}
