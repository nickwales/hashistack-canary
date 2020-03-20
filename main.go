package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/vault/api"
)

type HealthStatus struct {
	Description string
	Status      bool
}

var vaultClient *api.Client

var overallHealth map[string]HealthStatus

var startTime time.Time

func init() {
	overallHealth = make(map[string]HealthStatus)

	startTime = time.Now()

	vaultToken := os.Getenv("VAULT_TOKEN")
	if len(vaultToken) == 0 {
		fmt.Println("There's no Vault Token in the environment, exiting")
		os.Exit(1)
	}

	vaultAddr := os.Getenv("VAULT_ADDR")
	if len(vaultAddr) == 0 {
		vaultAddr = "http://127.0.0.1:8200"
	}

	var err error
	vaultClient, err = api.NewClient(&api.Config{Address: vaultAddr, HttpClient: nil})
	if err != nil {
		fmt.Println("Client errored: ", err)
	}

	vaultClient.SetToken(vaultToken)
}

func main() {

	h := make(chan map[string]HealthStatus, 5)

	go receiveStatuses(h)
	go getVaultStatus(h)

	http.HandleFunc("/alive.txt", KillMe)
	http.HandleFunc("/", GetHealthStatus)
	http.ListenAndServe(":8080", nil)
}

func receiveStatuses(h chan (map[string]HealthStatus)) {
	for {
		status := <-h
		fmt.Println("Received a status", status)
		for k, v := range status {
			overallHealth[k] = v
		}

	}

}

// GetHealthStatus returns status on all canary test statuses
func GetHealthStatus(w http.ResponseWriter, r *http.Request) {

	for _, status := range overallHealth {
		if status.Status == false {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	jsonString, _ := json.Marshal(overallHealth)

	fmt.Fprintf(w, "%s", jsonString)
}

// KillMe kills the canary after a specified time period
func KillMe(w http.ResponseWriter, r *http.Request) {

	var status string
	currentTime := time.Now()
	diff := currentTime.Sub(startTime)

	var deathTime float64
	deathTime = 1

	if diff.Minutes() > deathTime {
		fmt.Println("Our time is now")
		w.WriteHeader(http.StatusInternalServerError)
		status = "going down"
	} else {
		fmt.Println("Its not yet thats fosho")
		status = "staying Alive"
	}

	fmt.Fprintf(w, "Time alive: %s, we're %s", diff, status)

}
