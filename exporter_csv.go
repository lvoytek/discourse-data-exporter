package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"time"
)

var (
	CSVFoldername string
)

func SetCSVFolder(foldername string) error {
	CSVFoldername = foldername
	return os.Mkdir(foldername, 0755)
}

func ExportUsersCSV(users []UserEntry) {
	err := exportArrayToCSV("users.csv", users)

	if err != nil {
		log.Printf("ExportUsersCSV error: %v", err)
	}
}

func ExportTopicCommentsCSV(topicComments []TopicCommentsEntry) {
	err := exportArrayToCSV("topic_comments.csv", topicComments)

	if err != nil {
		log.Printf("ExportTopicCommentsCSV error: %v", err)
	}
}

func ExportTopicEditsCSV(topicEdits []TopicEditsEntry) {
	err := exportArrayToCSV("topic_edits.csv", topicEdits)

	if err != nil {
		log.Printf("ExportTopicEditsCSV error: %v", err)
	}
}

func exportArrayToCSV[T any](filename string, dataSet []T) error {
	if len(dataSet) == 0 {
		return nil
	}

	csvFile, err := os.Create(filepath.Join(CSVFoldername, filename))

	if err != nil {
		return err
	}

	defer csvFile.Close()

	writer := csv.NewWriter(csvFile)
	defer writer.Flush()

	// Get csv headers and write to file
	dataFields := reflect.TypeOf(dataSet[0])
	csvHeaders := []string{}

	for i := 0; i < dataFields.NumField(); i++ {
		csvHeaders = append(csvHeaders, dataFields.Field(i).Tag.Get("csv"))
	}

	err = writer.Write(csvHeaders)

	if err != nil {
		return err
	}

	// Write data to the file
	for _, nextEntry := range dataSet {
		nextEntryStrings := []string{}
		fields := reflect.ValueOf(nextEntry)
		for i := 0; i < fields.NumField(); i++ {
			field := fields.Field(i)

			// Convert each field to string based on its kind
			switch field.Kind() {
			case reflect.String:
				nextEntryStrings = append(nextEntryStrings, field.String())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				nextEntryStrings = append(nextEntryStrings, fmt.Sprintf("%d", field.Int()))
			case reflect.Bool:
				nextEntryStrings = append(nextEntryStrings, fmt.Sprintf("%t", field.Bool()))
			case reflect.Struct:
				if field.Type() == reflect.TypeOf(time.Time{}) {
					// Format time.Time as string
					nextEntryStrings = append(nextEntryStrings, field.Interface().(time.Time).Format(time.RFC3339))
				} else {
					nextEntryStrings = append(nextEntryStrings, "")
				}
			default:
				nextEntryStrings = append(nextEntryStrings, "")
			}
		}
		if err := writer.Write(nextEntryStrings); err != nil {
			return err
		}
	}

	return nil
}
