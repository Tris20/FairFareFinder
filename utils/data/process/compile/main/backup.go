
package main

import (
	"log"
	"os"
	"path/filepath"
	"time"
  "fmt"
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

		// Move existing results.db to backups folder with a timestamp
		timestamp := time.Now().Format("20060102_150405")
		backupPath := filepath.Join(backupDir, fmt.Sprintf("main_backup_%s.db", timestamp))
		err := os.Rename(dbPath, backupPath)
		if err != nil {
			log.Fatalf("Failed to move existing database to backup: %v", err)
		}
	}
}
