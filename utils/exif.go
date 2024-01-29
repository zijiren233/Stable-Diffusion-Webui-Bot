package utils

import (
	"errors"

	"github.com/barasher/go-exiftool"
)

var exif *exiftool.Exiftool

func init() {
	var err error
	exif, err = exiftool.NewExiftool()
	if err != nil {
		exif = nil
	}
}

func Exif() (*exiftool.Exiftool, error) {
	if exif == nil {
		return nil, errors.New("exiftool not found, please install and restart")
	}
	return exif, nil
}
