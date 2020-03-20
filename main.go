package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
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

var httpPort string

var lifeSupport int

func init() {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	lifeSupport = r1.Intn(10)

	fmt.Println("Lifesupport: ", lifeSupport)

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

	http.HandleFunc("/health/", KillMe)
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
	deathTime = 10 + float64(lifeSupport)

	if diff.Minutes() > deathTime {
		w.WriteHeader(http.StatusInternalServerError)
		status = "Its time to go down"
	} else {
		status = "Ooh ooh ooh, we're staying alive"
	}

	fmt.Fprintf(w, "Time alive: %s\nWe're staying up for: %v \n%s", diff, int(deathTime), status)
}
