package storage

import (
	"io"
	"os"
	"path/filepath"
)

// saves file to local storage
func SaveLocally(file io.Reader, fileName string) (string, error) {
	savePath := "uploads/"
	os.MkdirAll(savePath, os.ModePerm) 

	filePath := filepath.Join(savePath, fileName)
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		return "", err
	}

	return "/uploads/" + fileName, nil
}
