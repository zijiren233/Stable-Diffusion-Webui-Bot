package utils

import (
	"github.com/barasher/go-exiftool"
)

var Exif *exiftool.Exiftool

func init() {
	var err error
	Exif, err = exiftool.NewExiftool()
	if err != nil {
		panic(err)
	}
}
