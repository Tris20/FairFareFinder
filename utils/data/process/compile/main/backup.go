package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

func backupDatabase(dbPath, outputDir string) {
	// Check if results.db exists
	if _, err := os.Stat(dbPath); err == nil {
		// Create /out/backups directory if it does not exist
		backupDir := filepath.Join(outputDir, "backups")
		if _, err := os.Stat(backupDir); os.IsNotExist(err) {
			err := os.Mkdir(backupDir, 0755)
			if err != nil {
				log.Fatalf("Failed to create backup directory %s: %v", backupDir, err)
			}
		}

		// Create a timestamped backup file path
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, fmt.Sprintf("main_backup_%s.db", timestamp))

		// Copy the database to the backup file
		err := copyFile(dbPath, backupPath)
		if err != nil {
			log.Fatalf("Failed to copy database to backup: %v", err)
		}
	}
}

// Helper function to copy the file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Optionally, you can sync the destination file to ensure the content is written to disk
	err = destFile.Sync()
	return err
}
