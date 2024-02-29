package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
)

var mu sync.Mutex

func generateCodes(n int) []string {
	codes := make([]string, n)
	for i := 0; i < n; i++ {
		codes[i] = uuid.New().String()
	}
	return codes
}

func writeCodes() {
	mu.Lock()
	defer mu.Unlock()

	codes, err := json.Marshal(generateCodes(20))
	if err != nil {
		panic("could not generate codes")
	}

	if os.WriteFile("codes.json", codes, 0644) != nil {
		panic("could not write to file")
	}
}

func redeem(code string) bool {
	mu.Lock()
	defer mu.Unlock()

	var codes []string

	data, err := os.ReadFile("codes.json")
	if err != nil {
		panic("could not read file")
	}

	if err = json.Unmarshal(data, &codes); err != nil {
		panic("could not unmarshal json")
	}

	var updatedCodes []string
	found := false
	for _, c := range codes {
		if c != code {
			updatedCodes = append(updatedCodes, c)
		} else {
			found = true
		}
	}

	ucBytes, err := json.Marshal(updatedCodes)
	if err != nil {
		panic("could not marshal json")
	}

	if os.WriteFile("codes.json", ucBytes, 0644) != nil {
		panic("could not write to file")
	}

	return found
}

func main() {
	// writeCodes()
	fmt.Println(redeem("eb0cf47a-8350-455d-98df-72f9f0d412c2"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path)
	})
	http.ListenAndServe(":80", nil)
}
