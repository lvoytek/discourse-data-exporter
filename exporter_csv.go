package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
)

var (
	CSVFoldername string
)

func SetCSVFolder(foldername string) error {
	CSVFoldername = foldername
	return os.Mkdir(foldername, 0755)
}

func createCSV(filename string) (writer *csv.Writer, err error) {
	csvFile, err := os.Create(filepath.Join(CSVFoldername, filename))

	if err != nil {
		return nil, err
	}

	csvFile.Close()
	writer = csv.NewWriter(csvFile)

	return writer, nil
}
