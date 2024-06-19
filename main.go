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

	//mux.HandleFunc("/grokit", func(w http.ResponseWriter, r *http.Request) {
	//	fmt.Println("received connection to / from", r.RemoteAddr)
	//	w.WriteHeader(http.StatusOK)
	//	r.Body.Close()
	//
	//	//b, err := io.ReadAll(r.Body)
	//	//if err != nil {
	//	//	logrus.Print("body read err", err)
	//	//}
	//	//fmt.Println("body:", string(b))
	//	sema <- struct{}{}
	//	count++
	//	fmt.Println("count", count)
	//	<-sema
	//})

	mux.HandleFunc("/endpoint1", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received connection to endpoint1 from", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)

		sema <- struct{}{}
		count++
		fmt.Println("count", count)
		<-sema
	})

	mux.HandleFunc("/endpoint2", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("received connection to endpoint2 from", r.RemoteAddr)
		w.WriteHeader(http.StatusOK)

		sema <- struct{}{}
		count++
		fmt.Println("count", count)
		<-sema
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
