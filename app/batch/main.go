package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

func main() {
	log.Println("start job!!")
	db := dbConnect()
	defer db.Close()

	err := verifyNotification(db)
	if err != nil {
		log.Fatal("failed to update: ", err)
	}

	log.Println("job completed!!")
}

// dbConnect ...
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

// verifyNotification ... 通知を確認済みに更新する
// UsersテーブルとNotificationテーブルを結合し、Usersテーブルに存在するEmailアドレスの通知のみを確認済みに更新する
func verifyNotification(db *sql.DB) error {
	_, err := db.Exec(""+
		"UPDATE notification n "+
		"set verification = $1 "+
		"FROM users u WHERE n.email = u.email "+
		"AND verification = $2",
		true, false)
	if err != nil {
		return err
	}

	return nil
}
