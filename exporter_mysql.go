package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
)

var (
	mysqlDB *sql.DB
)

func ConnectMySQL(serverURL string, username string, password string) error {
	mysqlCfg := mysql.Config{
		User:   username,
		Passwd: password,
		Net:    "tcp",
		Addr:   serverURL,
		DBName: "discourse",
	}

	var err error
	mysqlDB, err = sql.Open("mysql", mysqlCfg.FormatDSN())

	if err != nil {
		return fmt.Errorf("mysql connection setup error: %v", err)
	}

	mysqlDB.SetConnMaxLifetime(time.Minute * 3)
	mysqlDB.SetMaxOpenConns(10)
	mysqlDB.SetMaxIdleConns(10)

	err = mysqlDB.Ping()

	if err != nil {
		return fmt.Errorf("mysql database ping error: %v", err)
	}

	return nil
}

func InitializeMySQLDatabase() error {
	// Users
	_, err := mysqlDB.Exec("CREATE TABLE IF NOT EXISTS users " +
		"(" +
		"user_id INT PRIMARY KEY, " +
		"username VARCHAR(120) UNIQUE NOT NULL, " +
		"name VARCHAR(120), " +
		"primary_group_name VARCHAR(120), " +
		"UNIQUE KEY idx_username (username)" +
		")")

	if err != nil {
		return fmt.Errorf("users table creation error: %v", err)
	}

	// Topic comments
	_, err = mysqlDB.Exec("CREATE TABLE IF NOT EXISTS comments " +
		"(" +
		"post_id INT PRIMARY KEY, " +
		"category_slug TEXT NOT NULL, " +
		"topic_id INT NOT NULL, " +
		"creation_time DATETIME NOT NULL, " +
		"update_time DATETIME NOT NULL, " +
		"username VARCHAR(120) NOT NULL, " +
		"is_initial_post BOOL NOT NULL, " +
		"CONSTRAINT fk_username_comments FOREIGN KEY (username) REFERENCES users(username)" +
		")")

	if err != nil {
		return fmt.Errorf("comments table creation error: %v", err)
	}

	// Topic edits
	_, err = mysqlDB.Exec("CREATE TABLE IF NOT EXISTS edits " +
		"(" +
		"topic_id INT, " +
		"edit_number INT, " +
		"creation_time DATETIME NOT NULL, " +
		"username VARCHAR(120) NOT NULL, " +
		"primary key (topic_id, edit_number), " +
		"CONSTRAINT fk_username_edits FOREIGN KEY (username) REFERENCES users(username)" +
		")")

	if err != nil {
		return fmt.Errorf("edits table creation error: %v", err)
	}

	return nil
}

func ExportUsersMySQL(users []UserEntry) {
	for _, user := range users {
		_, err := mysqlDB.Exec("INSERT INTO users "+
			"(user_id, username, name, primary_group_name) "+
			"VALUES (?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE "+
			"username = VALUES(username), "+
			"name = VALUES(name), "+
			"primary_group_name = VALUES(primary_group_name)",
			user.user_id, user.username, user.name, user.primary_group_name)
		if err != nil {
			log.Printf("ExportUsersMySQL error: %v", err)
		}
	}
}

func ExportTopicCommentsMySQL(topicComments []TopicCommentsEntry) {
	for _, topicComment := range topicComments {
		_, err := mysqlDB.Exec("INSERT INTO comments "+
			"(category_slug, topic_id, post_id, creation_time, update_time, username, is_initial_post) "+
			"VALUES (?, ?, ?, ?, ?, ?, ?) "+
			"ON DUPLICATE KEY UPDATE "+
			"update_time = VALUES(update_time)",
			topicComment.category_slug, topicComment.topic_id, topicComment.post_id, topicComment.creation_time, topicComment.update_time, topicComment.username, topicComment.is_initial_post)
		if err != nil {
			log.Printf("ExportTopicCommentsMySQL error: %v", err)
		}
	}
}

func ExportTopicEditsMySQL(topicEdits []TopicEditsEntry) {
	for _, topicEdit := range topicEdits {
		_, err := mysqlDB.Exec("INSERT IGNORE INTO edits (topic_id, edit_number, creation_time, username) VALUES (?, ?, ?, ?)",
			topicEdit.topic_id, topicEdit.edit_number, topicEdit.creation_time, topicEdit.username)
		if err != nil {
			log.Printf("ExportTopicEditsMySQL error: %v", err)
		}
	}
}
