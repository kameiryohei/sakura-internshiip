package backend

import (
	"encoding/json"
	"log"
	"fmt"
	"net/http"
	"os"
)

// JSON type: represents the structure of the incoming JSON data.
type JSON struct {
	Devices map[string]struct {
		MAC struct {
			Key string `json:"key"`
		}
		IP struct {
			Key string `json:"key"`
		}
		Vendor struct {
			Key string `json:"key"`
		}
	}
}


// RunBackend function: starts the HTTP server.
func RunBackend() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Register the handler function for the "/upload" endpoint.
	http.HandleFunc("/upload", uploadHandler)

	// Start the server on the specified port.
	log.Printf("Server listening on port %s...", port)

	// Handle any errors that occur while starting the server.
	err := http.ListenAndServe(":"+port, nil)
	log.Fatal(err)
}


// uploadHandler function: handles file uploads.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST.
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Println("Invalid request method")

		return
	}

	// Check for the Authorization header.
	// if r.Header.Get("Authorization") != "abcde" {
	// 	http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
	// 	log.Println("Unauthorized access attempt")

	// 	return
	// }

	// Call the parseJSON function to handle the request.
	JSON_data := parseJSON(w, r)

	i := 0
	for k, v := range JSON_data.Devices {
		fmt.Printf("Device %d (key: %s):\n", i, k)
		fmt.Printf("  MAC: %s\n", v.MAC.Key)
		fmt.Printf("  IP: %s\n", v.IP.Key)
		fmt.Printf("  Vendor: %s\n", v.Vendor.Key)
		i++
	}
}


// parseJSON function: parses JSON requests.
func parseJSON(w http.ResponseWriter, r *http.Request) JSON {
	var data JSON
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error parsing JSON: %v", err)

		return JSON{}
	}

	return data
}
