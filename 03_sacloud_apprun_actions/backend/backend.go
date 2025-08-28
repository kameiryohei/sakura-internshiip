package backend


import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)


func RunBackend() {
	// Register the handler function for the "/upload" endpoint.
	http.HandleFunc("/upload", uploadHandler)

	// Start the server on port 8000.
	fmt.Println("Server listening on port 8000...")

	// Handle any errors that occur while starting the server.
	err := http.ListenAndServe(":8000", nil)
	log.Fatal(err)
}


func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)

		return
	}

	if r.Header.Get("Authorization") == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Call the JSONParser function to handle the request.
	JSONParser(w, r)
}

func JSONParser(w http.ResponseWriter, r *http.Request) {
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Parsed JSON: %+v\n", data)
}