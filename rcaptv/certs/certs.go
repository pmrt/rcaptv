package certs

import (
	"os"
	"path/filepath"
)

var wd string

func init() {
	dir, err := os.Getwd()
	if err != nil {
		panic("could not get current working dir: " + err.Error())
	}
	wd = dir
}

func Filename(certName string) string {
	return filepath.Join(wd, "x509", certName)
}
