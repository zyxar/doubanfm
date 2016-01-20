package main

import (
	"errors"
	"os"
	"path/filepath"
)

var homeDir string

func init() {
	homeDir = filepath.Join(os.Getenv("HOME"), ".Douban.FM")
}

func mkHomeDir() (err error) {
	exists, err := isDirExists(homeDir)
	if err != nil {
		return
	}
	if exists {
		return
	}
	return os.Mkdir(homeDir, 0755)
}

func isDirExists(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err == nil {
		if stat.IsDir() {
			return true, nil
		}
		return false, errors.New(path + " exists but is not a directory")
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
