package main

import (
	"database/sql"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/tools/go/cfg"
)

var (
	mysqlDB *sql.DB
)

func ConnectMySQL(serverURL string, username string, password string) error {
	mysqlCfg := mysql.Config{
		User:      username,
		Passwd:    password,
		Net:       "tcp",
		Addr:      serverURL,
		DBName:    "discourse",
		parseTime: true,
	}

	var err error
	mysqlDB, err = sql.Open("mysql", cfg.FormatDSN())

	if err != nil {
		return err
	}

	mysqlDB.SetConnMaxLifetime(time.Minute * 3)
	mysqlDB.SetMaxOpenConns(10)
	mysqlDB.SetMaxIdleConns(10)

	err := db.Ping()
	return err
}

func InitializeMySQLDatabase() error {
	//Topic comments
	_, err := mysqlDB.exec("CREATE TABLE IF NOT EXISTS comments (post_id INT PRIMARY KEY, category_slug VARCHAR NOT NULL, topic_id INT NOT NULL, creation_time DATETIME NOT NULL, update_time DATETIME NOT NULL, username VARCHAR NOT NULL)")

	return err
}

func ExportTopicCommentsMySQL(topicComments []TopicCommentsEntry) {
	for _, topicComment := range topicComments {
		_, err := mysqlDB.exec("INSERT INTO comments (category_slug, topic_id, post_id, creation_time, update_time, username) VALUES (?, ?, ?, ?, ?, ?)",
			topicComment.category_slug, topicComment.topic_id, topicComment.post_id, topicComment.creation_time, topicComment.update_time, topicComment.username)
		if err != nil {
			log.Printf("ExportTopicCommentsMySQL error: %v", err)
		}
	}
}
