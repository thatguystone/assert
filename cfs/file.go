package cfs

import (
	"io"
	"io/ioutil"
	"os"
)

// Create creates the given file by ensuring that is parent directories exist,
// then it creates the file.
func Create(path string) (*os.File, error) {
	err := CreateParents(path)
	if err != nil {
		return nil, err
	}

	return os.Create(path)
}

// Write creates the file at the given path, creating parent directories as
// necessary.
func Write(path string, c []byte) error {
	err := CreateParents(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, c, 0640)
}

// Copy copies a file from src to dst.
func Copy(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}

	defer s.Close()

	err = CreateParents(dst)
	if err != nil {
		return err
	}

	d, err := os.Create(dst)
	if err == nil {
		defer d.Close()
		_, err = io.Copy(d, s)
	}

	return err
}
