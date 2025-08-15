package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	outputFile := "scenario_file_delete_m.csv"

	// Just hardcode the file paths here
	targetFiles := []string{
		"customer_m/cluster_0001/file_0001.txt",
		"customer_n/cluster_0001/file_0001.txt",
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Failed to create CSV:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, path := range targetFiles {
		record := []string{
			path,
			"delete",
			fmt.Sprintf("delete %s", path),
		}
		writer.Write(record)
	}

	fmt.Println("scenario_file_delete_m.csv created successfully.")
}