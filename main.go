package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var urls = make(map[string]string)

func main_page(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
		http.Redirect(w, r, "/shorten", http.StatusSeeOther)
		return
	}

	url_key := strings.TrimPrefix(r.URL.Path, "")
	url_key = strings.Replace(url_key, "/", "", -1)
	if r.Method == http.MethodGet && url_key != "" && url_key != "favicon.ico" {
		url_redirect(w, r)
	}

	w.Header().Set("Content-Type", "text/html")

	fmt.Fprint(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>URL Shortener</title>
		</head>
		<body>
			<h2>URL Shortener</h2>
			<form method="post" action="/shorten">
				<input type="url" name="url" placeholder="Enter url" required>
				<input type="submit" value="Shorten">
				</form>
				</body>
				</html>
	`)
}

func final_page(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	url_to_short := r.FormValue("url")
	if url_to_short == "" {
		http.Error(w, "The url cannot be determined", http.StatusBadRequest)
		return
	}

	url_key := key_generator()
	urls[url_key] = url_to_short

	connStr := "user=psqluser password=qwerty dbname=testtz sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Exec("insert into urls (url_key, url_long) values ($1, $2)", url_key, url_to_short)

	shortenedURL := fmt.Sprintf("http://localhost:8080/%s", url_key)

	w.Header().Set("Content-Type", "text/html")

	fmt.Fprint(w, `
	<!DOCTYPE html>
	<html>
	<head>
			<title>URL Shortener</title>
			</head>
		<body>
			<h2>URL Shortener</h2>
			<p>Начальный url: `, url_to_short, `</p>
			<p>Сокращенный url: <a href="`, shortenedURL, `">`, shortenedURL, `</a></p> 
		</body>
		</html>
		`) // TODO: Возможно, стоит либо убрать ссылку, либо сделать ссылки везде
}

func url_redirect(w http.ResponseWriter, r *http.Request) {
	url_key := strings.TrimPrefix(r.URL.Path, "/")
	if url_key == "" {
		http.Error(w, "The short key to the url is lost", http.StatusBadRequest)
		return
	}

	url_to_short, found := urls[url_key]
	if !found {
		http.Error(w, "The short key to the url is lost", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, url_to_short, http.StatusMovedPermanently)
}

func key_generator() string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const key_len = 7

	rand.Seed(time.Now().UnixNano())
	url_key := make([]byte, key_len)
	for i := range url_key {
		url_key[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(url_key)
}

func main() {
	http.HandleFunc("/", main_page)
	http.HandleFunc("/shorten", final_page)

	fmt.Println("Start on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}
