package main
import (
	"fmt"
	"strings"
	"database/sql"
	_ "modernc.org/sqlite"  //registers itself in the database/sql as a driver, did not need it's function,  so use _
	"crypto/sha256"
	"github.com/jxskiss/base62"
	"net/http"
	"log"
	"encoding/json"
)

func main(){
	initDB()
	serveFiles()
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/", redirectHandler)

	fmt.Println("Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Response struct{
	ShortURL     string `json:"short_url,omitempty"`
	OriginalLink string `json:"original_link,omitempty"`
	
}

func process_url(url string) string{
	No_http := strings.ReplaceAll(url, "https://", "")
	No_http = strings.ReplaceAll(No_http, "http://", "")
	url_hash := sha256.New()
	url_hash.Write([]byte(No_http))

	hashByte := url_hash.Sum(nil)
	encode_62 := base62.EncodeToString(hashByte[:5]) 
	return encode_62
}


func initDB(){
	db, err := sql.Open("sqlite", "urls.db") 
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()
	createTable := `
		CREATE TABLE IF NOT EXISTS urls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		original_link TEXT NOT NULL,
		code TEXT NOT NULL
	);`

	_, err = db.Exec(createTable)
	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}
}

func insertDB(original string, code string) error {
	db, err := sql.Open("sqlite", "urls.db")
	defer db.Close()
	if err != nil{
		fmt.Println("Error opening database:", err)
		return err
	}
	_, err = db.Exec("INSERT INTO urls (original_link, code) VALUES (?, ?)", original, code)
	if err != nil{
		fmt.Println("Error Inserting:", err)
		return err
	}
	return err
}


func SelectDB(code string) (string, error){
	db, err := sql.Open("sqlite", "urls.db")
	defer db.Close()
	if err != nil{
		return "", err
	}
	var OriginalLink string
	err = db.QueryRow("SELECT original_link FROM urls WHERE code = ?", code).Scan(&OriginalLink)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", err
		}
		return "", err
	}
	return OriginalLink, nil
}


func Formatter(code string) string{
	new_url := "http://cut.ly/"
	new_url += code
	return new_url
}

func shortenHandler(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	original := r.Form.Get("url")
	if original == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}
	code := process_url(original)

	err := insertDB(original, code)
	if err != nil{
		http.Error(w, "Something went wrong. Please try again later.", http.StatusBadRequest)
		return
	}

	shortURL := "http://localhost:8080/" + code
	json.NewEncoder(w).Encode(Response{ShortURL: shortURL})
}


func serveFiles(){
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
}


func redirectHandler(w http.ResponseWriter, r *http.Request){
	code := strings.TrimPrefix(r.URL.Path, "/")
    if code == "" {
        http.Error(w, "Short URL not found", http.StatusNotFound)
        return
    }
	originalURL, err := SelectDB(code)
	if err != nil{
		http.Error(w, "code not found", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusFound)
}