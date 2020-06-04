package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

func main() {
	serverPort, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	targetServer := getEnv("TARGET_SERVER", "http://127.0.0.1:8080/env")

	envVariables := make(map[string]string)
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "CUSTOM") {
			splitEnv := strings.Split(v, "=")
			envVariables[splitEnv[0]] = strings.Join(splitEnv[1:], "=")
		}
	}

	http.Handle("/env", handlers.LoggingHandler(log.StandardLogger().Out, MakeEnvHandler(envVariables)))
	//http.HandleFunc("/env", MakeEnvHandler(envVariables))
	http.Handle("/health", handlers.LoggingHandler(log.StandardLogger().Out, MakeHealthHandler()))
	http.Handle("/forward", handlers.LoggingHandler(log.StandardLogger().Out, MakeForwardHandler(targetServer)))
	log.Infof("Start listening on port %d", serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil))
}

func getEnv(key string, fallback string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}

	return value
}

func MakeEnvHandler(envVariables map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(envVariables)
	}
}

func MakeHealthHandler() http.HandlerFunc {
	healthyStatus := true
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			targetStatus := r.URL.Query().Get("healthy")
			healthyStatus = targetStatus == "true"
			w.WriteHeader(http.StatusOK)
		} else {
			if healthyStatus {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}
	}
}

func MakeForwardHandler(target string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(target)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}

		w.WriteHeader(resp.StatusCode)
		respBody, err := ioutil.ReadAll(resp.Body)
		w.Write(respBody)
	}
}
