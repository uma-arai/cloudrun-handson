package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type Response struct {
	Status  int
	Message string
}

func main() {
	var db *sql.DB

	if os.Getenv("DB_HOST") != "" {
		db = dbConnect()
		defer db.Close()
	}

	http.HandleFunc("/backend", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Referer())
		username := "backend"
		body := Response{http.StatusOK, "Hello World, " + username + "!"}
		res, err := json.Marshal(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	})

	http.HandleFunc("/notification", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.Referer())
		id := r.FormValue("id")
		msg := "no message"

		if id != "" {
			fmt.Println("id:", id)
			n, err := getNotification(db, id)
			if err != nil {
				log.Fatal(err)
			}
			byteMsg, err := json.Marshal(n)
			if err != nil {
				log.Fatal(err)
			}
			msg = string(byteMsg)
		}

		body := Response{http.StatusOK, msg}
		res, err := json.Marshal(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(res)
	})

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "healthcheck OK")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}

func dbConnect() *sql.DB {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("failed to init database: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	log.Default().Println("success to connect db!!")

	return db
}

type Notification struct {
	ID           string
	CreatedAt    string
	UpdatedAt    string
	IsRead       bool
	IsDeleted    bool
	Verification bool
	Email        string
	Body         string
}

func getNotification(db *sql.DB, id string) (*Notification, error) {
	n := &Notification{}
	err := db.QueryRow("select * from notification where id=$1", id).Scan(&n.ID, &n.CreatedAt, &n.UpdatedAt, &n.IsRead, &n.IsDeleted, &n.Verification, &n.Email, &n.Body)
	if err != nil {
		n.ID = id
		n.Body = "No message found"
	}

	return n, nil
}
