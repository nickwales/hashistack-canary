package main

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/api"
)

// Check the given secret is what we are expecting
// Check the given Vault token is:
//	 - still alive /auth/token/lookup-self
// 	 - has the appropriate policies

// TokenStatus checks the status of the current token
func TokenStatus() (map[string]HealthStatus, map[string]HealthStatus) {
	var tokenStatus, policyAvailable bool
	tokenInfo, err := vaultClient.Auth().Token().LookupSelf()
	if err != nil {
		fmt.Println("Bad vault, something went wrong getting out token: ", err)
		tokenStatus = false
		policyAvailable = false
	} else {
		tokenStatus = true
		policyAvailable = checkTokenPolicies(tokenInfo, "root")
	}

	a := HealthStatus{Description: "Checks the Vault token has not expired", Status: tokenStatus}
	tokenHealth := make(map[string]HealthStatus)
	tokenHealth["Token Health"] = a
	// fmt.Println(tokenHealth)
	// overallHealth["Vault Token Status"] = tokenHealth

	policyStatus := make(map[string]HealthStatus)
	policyStatus["Vault Token Policy"] = HealthStatus{Description: "Checks the Vault token has the right policies", Status: policyAvailable}
	// fmt.Println(policyHealth)
	// overallHealth["Vault Token Policy"] = policyHealth

	return tokenHealth, policyStatus
}

func checkTokenPolicies(token *api.Secret, policy string) (policyAvailable bool) {
	requiredPolicy := "hashistack-canary"
	policies := token.Data["policies"]
	for _, policy := range policies.([]interface{}) {
		if policy == requiredPolicy {
			policyAvailable = true
		} else {
			policyAvailable = false
		}
	}
	return policyAvailable
}

func getVaultStatus(h chan map[string]HealthStatus) {
	for {
		tokenHealth, policyStatus := TokenStatus()
		h <- tokenHealth
		h <- policyStatus
		time.Sleep(10 * time.Second)
	}
}
