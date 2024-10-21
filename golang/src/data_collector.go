package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type NewsArticle struct {
	Title string `json:"title"`
	Body  string `json:"body"`
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

	var articles []NewsArticle
	for _, article := range result["articles"].([]interface{}) {
		art := article.(map[string]interface{})
		articles = append(articles, NewsArticle{
			Title: art["title"].(string),
			Body:  art["description"].(string),
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
