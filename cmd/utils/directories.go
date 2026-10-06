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
	"chats",
	"groups",
	"groups/avatars",
}

const DEFAULT_PATH = "./images/avatar/default.png"

func MakeUploadDirectories() error {
	if err := os.MkdirAll("./uploads", os.ModePerm); err != nil {
		return err
	}

	for _, d := range UPLOADS_DIRECTORIES {
		if err := os.MkdirAll(filepath.Join("./uploads", d), os.ModePerm); err != nil {
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

var SeedMap = []struct{ Src, Dst string }{
	{"./images/samples/posts", "./uploads/posts"},
	{"./images/samples/avatars", "./uploads/avatars"},
}

func SeedUploads() error {
	for _, m := range SeedMap {
		if err := CopyDir(m.Src, m.Dst); err != nil {
			return fmt.Errorf("seeding %s -> %s: %w", m.Src, m.Dst, err)
		}
	}
	return nil
}

func CopyDir(src, dst string) error {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil
	}

	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, os.ModePerm)
		}

		if _, err := os.Stat(target); err == nil {
			return nil
		}

		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}