package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func main() {
	outputFile := "scenario_create-update_o.csv"
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Failed to create CSV file:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for dirNum := 1; dirNum <= 10; dirNum++ {
		for fileNum := 1; fileNum <= 10; fileNum++ {
			dir := fmt.Sprintf("customer_o/cluster_%04d", dirNum)
			fileName := fmt.Sprintf("file_%04d.txt", fileNum)
			filePath := fmt.Sprintf("%s/%s", dir, fileName)

			// Create row: 3 columns (no content needed)
			writer.Write([]string{
				filePath,
				"create",
				"initial commit",
			})

			// Update row: 4 columns (content = "test data")
			writer.Write([]string{
				filePath,
				"update",
				fmt.Sprintf("update %s", fileName),
				"test data",
			})
		}
	}

	fmt.Println("✅ scenario_create-update_o.csv has been generated successfully.")
}