package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"flag"
	"fmt"
	"hash"
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

func getFileDetails(filePath string, hasherInstance getHasherInstance) (string, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening %s \n", err)
		return "", 0, err
	}
	defer file.Close()

	hasher := hasherInstance.getInstance()
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

func processFolder(filepath *string, hasherInstance getHasherInstance) map[string]*fileDescription {
	files := getFiles(*filepath)

	fmt.Println("Analysing files...")
	var fileLedger = map[string]*fileDescription{}
	for _, file := range files {
		fmt.Println(file)
		hash, fileSize, errorReading := getFileDetails(file, hasherInstance)
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
	return fileLedger
}

func displayResult(fileLedger map[string]*fileDescription) {
	var duplicateSpaceUsed int64
	var noOfDuplicateFiles int
	for key, value := range fileLedger {
		if value.hasDuplicates() {
			noOfDuplicateFiles++
			duplicateSpaceUsed += value.getSize()
			fmt.Println("Duplicate file found:")
			fmt.Printf("Hash Value: %s\n", key)
			fmt.Printf("File Size: %d bytes\n", value.getSize())
			fmt.Println("Duplicate Files:")
			for _, fileName := range value.getFileNames() {
				fmt.Println(fileName)
			}
			fmt.Println()
		}
	}

	if noOfDuplicateFiles > 0 {
		fmt.Printf("Number of duplicate files: %d\n", noOfDuplicateFiles)
		fmt.Printf("Total space wasted: %d bytes.\n", duplicateSpaceUsed)
	} else {
		fmt.Println("No Duplicate files found")
	}
}

type getHasherInstance interface {
	getInstance() hash.Hash
}

type sha256Instance struct {
}

type sha512Instance struct {
}

type md5Instance struct {
}

func (instance sha256Instance) getInstance() hash.Hash {
	return sha256.New()
}

func (instance md5Instance) getInstance() hash.Hash {
	return md5.New()
}

func (instance sha512Instance) getInstance() hash.Hash {
	return sha512.New()
}

func getHasher(hashName string) (getHasherInstance, bool) {
	switch hashName {
	case "sha256":
		return sha256Instance{}, true
	case "sha512":
		return sha512Instance{}, true
	case "md5":
		return md5Instance{}, true
	default:
		fmt.Println("Hash is not supported.")
		return nil, false
	}
}

func main() {
	filepath := flag.String("dir", ".", "the directory for finding duplicate files")
	hashAlgo := flag.String("crypto", "sha256", "the hash method to use")
	flag.Parse()

	hasherInstance, found := getHasher(*hashAlgo)
	if found {
		fileLedger := processFolder(filepath, hasherInstance)
		displayResult(fileLedger)
	}
}
