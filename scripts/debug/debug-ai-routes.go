package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Test direct access to backend without proxy
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			Proxy: nil, // Disable proxy
		},
	}

	// Test endpoints
	endpoints := []string{
		"http://localhost:8080/api/v1/ai/daily-inspiration",
		"http://localhost:8080/api/ai/daily-inspiration",
		"http://localhost:8080/api/v1/ai/stats",
		"http://localhost:8080/api/ai/stats",
	}

	fmt.Println("Testing AI endpoints...")
	fmt.Println("======================")

	for _, endpoint := range endpoints {
		fmt.Printf("\nTesting: %s\n", endpoint)
		
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		fmt.Printf("Status: %d %s\n", resp.StatusCode, resp.Status)
		
		if resp.StatusCode != 200 {
			// Read error body
			body := make([]byte, 1000)
			n, _ := resp.Body.Read(body)
			fmt.Printf("Response: %s\n", string(body[:n]))
		} else {
			fmt.Println("Success!")
		}
	}
}