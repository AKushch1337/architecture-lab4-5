package main

import (
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/AKushch1337/architecture-lab4-5/httptools"
	"github.com/AKushch1337/architecture-lab4-5/signal"
)

var port = flag.Int("port", 8080, "server port")
var test = flag.Bool("test", false, "if yes, server health status can be modified for test purpose")

const confResponseDelaySec = "CONF_RESPONSE_DELAY_SEC"
const confHealthFailure = "CONF_HEALTH_FAILURE"

func main() {
	h := new(http.ServeMux)

	healthCorrupt := false
	if *test {
		h.HandleFunc("/health-corrupt", func(rw http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				http.Error(rw, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			healthCorrupt = true
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		})
	}

	h.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("content-type", "text/plain")
		if failConfig := os.Getenv(confHealthFailure); failConfig == "true" || healthCorrupt {
			rw.WriteHeader(http.StatusInternalServerError)
			_, _ = rw.Write([]byte("FAILURE"))
		} else {
			rw.WriteHeader(http.StatusOK)
			_, _ = rw.Write([]byte("OK"))
		}
	})

	report := make(Report)

	h.HandleFunc("/api/v1/some-data", func(rw http.ResponseWriter, r *http.Request) {
		respDelayString := os.Getenv(confResponseDelaySec)
		if delaySec, parseErr := strconv.Atoi(respDelayString); parseErr == nil && delaySec > 0 && delaySec < 300 {
			time.Sleep(time.Duration(delaySec) * time.Second)
		}

		report.Process(r)

		rw.Header().Set("content-type", "application/json")
		rw.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(rw).Encode([]string{
			"1", "2",
		})
	})

	h.Handle("/report", report)

	server := httptools.CreateServer(*port, h)
	server.Start()
	signal.WaitForTerminationSignal()
}
