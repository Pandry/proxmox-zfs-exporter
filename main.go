package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getProxmoxCredentials() *ProxmoxAPI {
	config := ProxmoxAPI{}
	// Inserting "reasonable" defaults
	defaultHost := "127.0.0.1"
	defaultPort := "8006"
	defaultUser := "root"
	defaultPass := "password"
	config.Host = defaultHost
	config.Port = defaultPort
	config.User = defaultUser
	config.Pass = defaultPass

	f, err := os.Open("/etc/proxmox-zfs-exporter/config.json")
	if err != nil {
		log.Println("[WARN] Cannot open config file. Attempting read from environment...")
	} else {
		defer f.Close()

		enc := json.NewDecoder(f)
		err = enc.Decode(&config)
		if err != nil {
			log.Panic("[ERROR] Cannot decode config file.")
		}
	}

	val, exists := os.LookupEnv("PROX_USER")
	if exists {
		config.User = val
	}

	val, exists = os.LookupEnv("PROX_PASS")
	if exists {
		config.Pass = val
	}

	val, exists = os.LookupEnv("PROX_HOST")
	if exists {
		config.Host = val
	}

	val, exists = os.LookupEnv("PROX_PORT")
	if exists {
		config.Port = val
	}

	if defaultHost == config.Host {
		log.Println("[WARN] Using the default host (" + defaultHost + ")")
	}
	if defaultPass == config.Pass {
		log.Println("[WARN] Using the default password (" + defaultPass + ")")
	}
	if defaultUser == config.User {
		log.Println("[WARN] Using the default user (" + defaultUser + ")")
	}

	return &config
}

func main() {
	proxmoxAPI := getProxmoxCredentials()
	collector := newProxmoxZpoolCollector("prox_exporter", proxmoxAPI)
	prometheus.MustRegister(collector)

	go proxmoxAPI.refreshTicket()
	//Wait for the first ticket to be set
	proxmoxAPI.waitForTicket()

	http.Handle("/metrics", promhttp.Handler())
	listenPort, exists := os.LookupEnv("PORT")
	if !exists {
		listenPort = "9000"
	}
	log.Fatal(http.ListenAndServe(":"+listenPort, nil))
}
