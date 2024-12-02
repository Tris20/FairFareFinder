package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	rootDir := "../pixabay/images/" // Set the source directory
	outputDir := "output"           // Set the output directory for zip files

	// Ensure the output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			fmt.Println("Error creating output directory:", err)
			return
		}
	}

	// Walk through the directory
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a directory and not the root itself
		if info.IsDir() && path != rootDir {
			// Construct the filename for the zip file
			relativePath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			zipFileName := filepath.Join(outputDir, relativePath+".zip")

			// Create zip file
			err = zipFolder(path, zipFileName)
			if err != nil {
				fmt.Println("Error zipping directory:", err)
			} else {
				fmt.Println("Zipped directory:", zipFileName)
			}
			return filepath.SkipDir // Skip subdirectories
		}
		return nil
	})

	if err != nil {
		fmt.Println("Error walking the directory:", err)
	}
}

func zipFolder(folder, zipFileName string) error {
	zipFile, err := os.Create(zipFileName)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = filepath.Join(filepath.Base(folder), strings.TrimPrefix(path, folder))

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}

		return err
	})

	return nil
}
