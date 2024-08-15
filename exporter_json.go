package main

import (
	"encoding/json"
	"fmt"
	"log"
)

func ExportJSON(data DataToExport, itemsToExport ItemsToExport) {
	if !itemsToExport.Users {
		data.Users = nil
	}

	if !itemsToExport.TopicComments {
		data.Posts = nil
	}

	if !itemsToExport.TopicEdits {
		data.Edits = nil
	}

	jsonData, err := json.Marshal(data)

	if err != nil {
		log.Printf("ExportJSON error: %v", err)
	}

	fmt.Println(string(jsonData))
}
