package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
)

type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func main() {
	http.HandleFunc("/frontend", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello cnsrun handson's user:D\n")
	})

	http.HandleFunc("/random", func(w http.ResponseWriter, r *http.Request) {
		// 30%の確率で500エラーを返す
		if rand.N(10) < 3 {
			http.Error(w, "Error", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "200 OK\n")
	})

	http.HandleFunc("/backend", func(w http.ResponseWriter, r *http.Request) {
		backend := os.Getenv("BACKEND_FQDN")
		id := r.FormValue("id")
		url := backend + "/backend?id=" + id
		audience := backend + "/"

		resp, err := makeGetRequest(w, url, audience)
		if err != nil {
			fmt.Println("Error Request:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("Error Response:", resp.Status)
			return
		}
	})

	http.HandleFunc("/backend/notification", func(w http.ResponseWriter, r *http.Request) {
		backend := os.Getenv("BACKEND_FQDN")
		id := r.FormValue("id")
		url := backend + "/notification?id=" + id
		audience := backend + "/"

		resp, err := makeGetRequest(w, url, audience)
		if err != nil {
			fmt.Println("Error Request:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			fmt.Println("Error Response:", resp.Status)
			return
		}
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "healthcheck OK")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

// `makeGetRequest` makes a request to the provided `targetURL`
// with an authenticated client using audience `audience`.
func makeGetRequest(w io.Writer, targetURL string, audience string) (*http.Response, error) {
	ctx := context.Background()

	client, err := idtoken.NewClient(ctx, audience)
	if err != nil {
		return nil, fmt.Errorf("idtoken.NewClient: %w", err)
	}
	fmt.Println("audience:", audience)
	fmt.Printf("client: %#v\n", client)

	resp, err := client.Get(targetURL)
	if err != nil {
		return nil, fmt.Errorf("client.Get: %w", err)
	}
	fmt.Printf("resp: %#v\n", resp)

	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return nil, fmt.Errorf("io.Copy: %w", err)
	}

	return resp, nil
}
