package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var UPLOADS_DIRECTORIES = []string{
	"avatars",
	"posts",
	"comments",
}

const DEFAULT_PATH = "./images/avatar/default.png"

func MakeUploadDirectories() error {
	if err := os.Mkdir("./uploads/", os.ModePerm); err != nil && !os.IsExist(err) {
		return err
	}

	for _, d := range UPLOADS_DIRECTORIES {
		if err := os.Mkdir("./uploads/"+d, os.ModePerm); err != nil && !os.IsExist(err) {
			return err
		}
	}

	return nil
}

func CopyDefaultAvatar() error {
	sourceFileStat, err := os.Stat(DEFAULT_PATH)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", DEFAULT_PATH)
	}

	source, err := os.Open(DEFAULT_PATH)
	if err != nil {
		return err
	}
	defer source.Close()

	destinationPath := filepath.Join("./uploads/avatars", filepath.Base(DEFAULT_PATH))

	destination, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
