package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func ExportUsersJSON(users []UserEntry) {
	data, err := json.Marshal(users)

	if err != nil {
		log.Printf("ExportUsersJSON error: %v", err)
	}

	fmt.Println(string(data))
}

func ExportTopicCommentsJSON(topicComments []TopicCommentsEntry) {
	data, err := json.Marshal(topicComments)

	if err != nil {
		log.Printf("ExportUsersJSON error: %v", err)
	}

	fmt.Println(string(data))
}

func ExportTopicEditsJSON(topicEdits []TopicEditsEntry) {
	data, err := json.Marshal(topicEdits)

	if err != nil {
		log.Printf("ExportUsersJSON error: %v", err)
	}

	fmt.Println(string(data))
}
