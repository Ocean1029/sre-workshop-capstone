package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var crashTotal = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_crash_total",
	Help: "Number of times the /crash endpoint was hit.",
})

func newMux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/crash", func(w http.ResponseWriter, _ *http.Request) {
		crashTotal.Inc()
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("crashed\n"))
	})

	mux.Handle("/metrics", promhttp.Handler())

	return mux
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, newMux()))
}
