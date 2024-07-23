package main

import (
	"database/sql"
	"fmt"
	"time"
)

func ConnectMySQL(serverURL string, username string, password string) error {
	mysql_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s", username, password, serverURL))

	mysql_db.SetConnMaxLifetime(time.Minute * 3)
	mysql_db.SetMaxOpenConns(10)
	mysql_db.SetMaxIdleConns(10)

	if err != nil {
		mysql_db = nil
	}

	return err
}

func ExportTopicCommentsMySQL(topicComments []TopicCommentsEntry) error {

	return nil
}
