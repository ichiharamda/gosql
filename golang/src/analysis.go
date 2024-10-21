package main

import (
	"database/sql"
	"fmt"
)

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
