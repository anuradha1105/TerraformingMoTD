package main

import (
"database/sql"
"encoding/json"
"fmt"
"log"
"net/http"
"os"
"time"

_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB
var token string

type updateRequest struct {
Token   string `json:"token"`
Message string `json:"message"`
}

func getenv(key, fallback string) string {
if v := os.Getenv(key); v != "" {
return v
}
return fallback
}

func main() {
dbHost := getenv("DB_HOST", "127.0.0.1")
dbPort := getenv("DB_PORT", "3306")
dbUser := getenv("DB_USER", "root")
dbPassword := os.Getenv("DB_PASSWORD")
dbName := getenv("DB_NAME", "motd")
token = os.Getenv("MOTD_TOKEN")
if token == "" {
log.Fatal("MOTD_TOKEN environment variable must be set")
}

dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

var err error
db, err = sql.Open("mysql", dsn)
if err != nil {
log.Fatalf("failed to open database: %v", err)
}
defer db.Close()

waitForDB()

http.HandleFunc("/", handleIndex)
http.HandleFunc("/message", handleMessage)

addr := ":8080"
log.Printf("motd web server listening on %s", addr)
log.Fatal(http.ListenAndServe(addr, nil))
}

func waitForDB() {
for i := 0; i < 30; i++ {
if err := db.Ping(); err == nil {
return
}
time.Sleep(1 * time.Second)
}
log.Fatal("could not connect to database after 30s")
}

func currentMessage() (string, error) {
var message string
row := db.QueryRow("SELECT message FROM message_of_the_day ORDER BY id DESC LIMIT 1")
if err := row.Scan(&message); err != nil {
return "", err
}
return message, nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

message, err := currentMessage()
if err != nil {
log.Printf("error reading message: %v", err)
http.Error(w, "internal server error", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "text/plain; charset=utf-8")
fmt.Fprintln(w, message)
}

func handleMessage(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
return
}

var req updateRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
http.Error(w, "invalid request body", http.StatusBadRequest)
return
}

if req.Token == "" || req.Token != token {
http.Error(w, "invalid token", http.StatusUnauthorized)
return
}

if req.Message == "" {
http.Error(w, "message must not be empty", http.StatusBadRequest)
return
}

_, err := db.Exec("UPDATE message_of_the_day SET message = ? WHERE id = 1", req.Message)
if err != nil {
log.Printf("error updating message: %v", err)
http.Error(w, "internal server error", http.StatusInternalServerError)
return
}

w.WriteHeader(http.StatusOK)
fmt.Fprintln(w, "ok")
}
