package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	outputFile := "scenario_folder_delete_k.csv"

	// List the folders you want to delete
	targetFolders := []string{
		"customer_k/cluster_0001",
		"customer_k/cluster_0002",
	}

	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Failed to create CSV:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, folder := range targetFolders {
		record := []string{
			folder,
			"delete",
			fmt.Sprintf("delete folder %s", folder),
		}
		writer.Write(record)
	}

	fmt.Println("scenario_folder_delete_k.csv created successfully.")
}