package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type NewsArticle struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func open(path string, count uint) *sql.DB {
	db, err := sql.Open("mysql", path)
	if err != nil {
		log.Fatal("open error:", err)
	}

	for count > 0 {
		if err = db.Ping(); err == nil {
			fmt.Println("db connected!!")
			return db
		}
		time.Sleep(time.Second * 2)
		count--
		fmt.Printf("retry... count:%v\n", count)
	}

	log.Fatal("failed to connect to database:", err)
	return nil
}

func connectDB() *sql.DB {
	var path string = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&parseTime=true",
		os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"),
		os.Getenv("MYSQL_HOST"), os.Getenv("MYSQL_PORT"),
		os.Getenv("MYSQL_DB"))

	return open(path, 100)
}

func fetchNews() ([]NewsArticle, error) {
	apiKey := os.Getenv("NEWS_API_KEY")
	url := fmt.Sprintf("https://newsapi.org/v2/top-headlines?country=us&apiKey=%s", apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	articlesData, ok := result["articles"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for articles")
	}

	var articles []NewsArticle
	for _, article := range articlesData {
		art, ok := article.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected type for article")
		}
		title, ok := art["title"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for title")
		}
		body, ok := art["description"].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for description")
		}
		articles = append(articles, NewsArticle{
			Title: title,
			Body:  body,
		})
	}

	return articles, nil
}

func saveArticles(db *sql.DB, articles []NewsArticle) error {
	for _, article := range articles {
		_, err := db.Exec("INSERT INTO article (title, body) VALUES (?, ?)", article.Title, article.Body)
		if err != nil {
			return err
		}
	}
	return nil
}

func analyzeArticles(db *sql.DB, keyword string) {
	rows, err := db.Query("SELECT title, body FROM article WHERE title LIKE ?", "%"+keyword+"%")
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var articles []NewsArticle
	for rows.Next() {
		article := NewsArticle{}
		err = rows.Scan(&article.Title, &article.Body)
		if err != nil {
			panic(err)
		}
		articles = append(articles, article)
	}

	fmt.Printf("Articles containing the keyword '%s':\n", keyword)
	for _, article := range articles {
		fmt.Printf("Title: %s\nBody: %s\n", article.Title, article.Body)
	}
}

func startWebServer(db *sql.DB) {
	http.HandleFunc("/articles", func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT title, body FROM article")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var articles []NewsArticle
		for rows.Next() {
			article := NewsArticle{}
			err = rows.Scan(&article.Title, &article.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			articles = append(articles, article)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(articles)
	})

	fmt.Println("Starting web server on port 8080")
	http.ListenAndServe(":8080", nil)
}

func main() {
	db := connectDB()
	defer db.Close()

	// データ収集
	articles, err := fetchNews()
	if err != nil {
		log.Fatal("failed to fetch news:", err)
	}

	// データ保存
	err = saveArticles(db, articles)
	if err != nil {
		log.Fatal("failed to save articles:", err)
	}

	// データ分析
	analyzeArticles(db, "keyword")

	// Webサーバーの起動
	startWebServer(db)
}
