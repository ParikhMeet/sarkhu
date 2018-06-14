package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fileDescription struct {
	filenames []string
	size      int64
}

func (f *fileDescription) addDuplicateFile(file string) {
	f.filenames = append(f.filenames, file)
}

func (f *fileDescription) hasDuplicates() bool {
	return len(f.filenames) > 1
}

func (f *fileDescription) getFileNames() []string {
	return f.filenames
}

func (f *fileDescription) getSize() int64 {
	return f.size
}

func getFileDetails(filePath string) (string, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening %s \n", err)
		return "", 0, err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		panic(err)
	}

	fileStat, err := file.Stat()
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), fileStat.Size(), nil
}

func getFiles(directory string) []string {
	fileList := []string{}
	err := filepath.Walk(directory, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return fileList
}

func main() {
	filepath := flag.String("dir", ".", "the directory for finding duplicate files")
	flag.Parse()

	files := getFiles(*filepath)

	fmt.Println("Analysing files...")
	var fileLedger = map[string]*fileDescription{}
	for _, file := range files {
		fmt.Println(file)
		hash, fileSize, errorReading := getFileDetails(file)
		if errorReading != nil {
			continue
		}
		if val, exists := fileLedger[hash]; exists {
			val.addDuplicateFile(file)
		} else {
			fileLedger[hash] = &fileDescription{[]string{file}, fileSize}
		}
		fmt.Println(hash + "\n") // Added for extra blank line
	}

	var duplicateSpaceUsed int64
	for key, value := range fileLedger {
		if value.hasDuplicates() {
			duplicateSpaceUsed += value.getSize()
			fmt.Println("Duplicate file found:")
			fmt.Printf("Hash Value : %s\n", key)
			fmt.Printf("File Size: %d bytes\n", value.getSize())
			fmt.Println("Duplicate Files:")
			for _, fileName := range value.getFileNames() {
				fmt.Println(fileName)
			}
			fmt.Println()
		}
	}

	if duplicateSpaceUsed > 0 {
		fmt.Printf("Total space wasted: %d bytes.\n", duplicateSpaceUsed)
	} else {
		fmt.Println("No Duplicate files found")
	}
}
