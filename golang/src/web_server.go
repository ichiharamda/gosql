package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
)

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
