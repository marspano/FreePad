package objects

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JustKato/FreePad/lib/helper"
)

type Post struct {
	Name         string `json:"name"`
	LastModified string `json:"last_modified"`
	Content      string `json:"content"`
}

// Get the path to the storage directory
func getStorageDirectory() string {

	baseStoragePath, exists := os.LookupEnv("PAD_STORAGE_PATH")
	if !exists {
		baseStoragePath = "/tmp/"
	}

	// Check if the base storage path exists
	if _, err := os.Stat(baseStoragePath); os.IsNotExist(err) {
		// Looks like the base storage path was NOT set, create the dir
		err = os.Mkdir(baseStoragePath, 0777)
		// Check for errors
		if err != nil {
			// No way this sends an error unless it goes horribly wrong.
			panic(err)
		}
	}

	// Return the base storage path
	return baseStoragePath
}

// Get a post from the file system
func GetPost(fileName string) Post {
	// Get the base storage directory and make sure it exists
	storageDir := getStorageDirectory()

	// Generate the file path
	filePath := fmt.Sprintf("%s%s", storageDir, fileName)

	p := Post{
		Name:         fileName,
		Content:      "",
		LastModified: "Never Before",
	}

	// Check if the file exits
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		// File does exist, read it and set the content
		data, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Println("Error:", err)
		}

		// Get the content of the file and put it in the response
		p.Content = string(data)

		// Get last modified date
		fileData, err := os.Stat(filePath)
		if err == nil {
			p.LastModified = fileData.ModTime().Format("02/01/2006 03:04:05 PM")
		} else {
			fmt.Println(err)
		}
	}

	return p
}

// Write a post to the file system
func WritePost(p Post) error {

	maximumPadSize := helper.GetMaximumPadSize()
	if len(p.Content) > maximumPadSize {
		return errors.New("The pad is too big, please limit to the maximum of " + fmt.Sprint(maximumPadSize) + " characters")
	}

	// Get the base storage directory and make sure it exists
	storageDir := getStorageDirectory()

	// Generate the file path
	filePath := fmt.Sprintf("%s%s", storageDir, p.Name)

	f, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	// Write the contnets
	_, err = f.WriteString(p.Content)
	if err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return nil
}

// Cleanup all of the older posts based on the environment settings
func CleanupPosts(age int) {
	// Initialize the files buffer
	var files []string

	// Get the base path
	root := getStorageDirectory()

	// Scan the directory
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		// Check that the file is not a dir
		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	// Check if we had any errors
	if err != nil {
		fmt.Println("[Task::Cleanup]: Error", err)
	}

	// The timestamp where the file should be deleted
	tooOldTime := time.Now()

	// Go through all files and process them
	for _, filePath := range files {

		// Get last modified date
		fileData, err := os.Stat(filePath)
		if err == nil {
			fileAge := fileData.ModTime()

			// Check if the file is too old
			if fileAge.Add(time.Duration(age)*time.Minute).Unix() < tooOldTime.Unix() {
				fmt.Println("[Task::Cleanup]: Removing File", filePath)

				// Remove the file
				err = os.Remove(filePath)
				if err != nil {
					fmt.Println("[Task::Cleanup]: Failed to remove file", filePath)
				}
			}

		} else {
			fmt.Println("[Task::Cleanup]: Error", err)
		}

	}
}
