package utils

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"
)

type Photo struct {
	Width  int
	Height int
	Type   string
	Bytes  []byte
}

func CompressImageResource(data []byte, quality uint) (*Photo, error) {
	var imgSrc image.Image
	dataType, err := GetType(data)
	if err != nil {
		return nil, err
	}
	rowPhoto := bytes.NewReader(data)
	switch dataType {
	case "image/png":
		imgSrc, err = png.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	case "image/jpeg":
		imgSrc, err = jpeg.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	default:
		imgSrc, _, err = image.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	}
	newImg := image.NewRGBA(imgSrc.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
	buf := bytes.NewBuffer(nil)
	err = jpeg.Encode(buf, newImg, &jpeg.Options{Quality: int(quality)})
	if err != nil {
		return nil, err
	}
	jp := new(Photo)
	jp.Type = "image/jpeg"
	jp.Width = newImg.Bounds().Max.X
	jp.Height = newImg.Bounds().Max.Y
	if buf.Len() > rowPhoto.Len() {
		jp.Bytes = data
	} else {
		jp.Bytes = buf.Bytes()
	}
	return jp, nil
}

func GetPhotoSize(photo []byte) (width, hight int, err error) {
	var imgSrc image.Image
	dataType, err := GetType(photo)
	if err != nil {
		return 0, 0, err
	}
	rowPhoto := bytes.NewReader(photo)
	switch dataType {
	case "image/png":
		imgSrc, err = png.Decode(rowPhoto)
		if err != nil {
			return 0, 0, err
		}
		return imgSrc.Bounds().Max.X, imgSrc.Bounds().Max.Y, nil
	case "image/jpeg":
		imgSrc, err = jpeg.Decode(rowPhoto)
		if err != nil {
			return 0, 0, err
		}
		return imgSrc.Bounds().Max.X, imgSrc.Bounds().Max.Y, nil
	default:
		imgSrc, _, err = image.Decode(rowPhoto)
		if err != nil {
			return 0, 0, err
		}
		return imgSrc.Bounds().Max.X, imgSrc.Bounds().Max.Y, nil
	}
}

func PhotoColorInvert(photo []byte) (InvertPhoto []byte, Err error) {
	var imgSrc image.Image
	dataType, err := GetType(photo)
	if err != nil {
		return nil, err
	}
	rowPhoto := bytes.NewReader(photo)
	switch dataType {
	case "image/png":
		imgSrc, err = png.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	case "image/jpeg":
		imgSrc, err = jpeg.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	default:
		imgSrc, _, err = image.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	}
	bounds := imgSrc.Bounds()
	newImg := image.NewRGBA(bounds)
	for x := bounds.Min.X; x < bounds.Max.X; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			originalColor := imgSrc.At(x, y)
			r, g, b, a := originalColor.RGBA()
			newColor := color.RGBA{uint8(255 - r), uint8(255 - g), uint8(255 - b), uint8(a)}
			newImg.Set(x, y, newColor)
		}
	}
	buffer := bytes.NewBuffer(nil)
	err = jpeg.Encode(buffer, newImg, &jpeg.Options{Quality: 100})
	return buffer.Bytes(), err
}

func CompressImageResourceToSize(data []byte, maxSize uint) (*Photo, error) {
	if maxSize == 0 {
		return nil, errors.New("max size is zero")
	}
	var imgSrc image.Image
	dataType, err := GetType(data)
	if err != nil {
		return nil, err
	}
	jp := new(Photo)
	rowPhoto := bytes.NewReader(data)
	switch dataType {
	case "image/png":
		imgSrc, err = png.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
		if rowPhoto.Size() < int64(maxSize) {
			jp.Width = imgSrc.Bounds().Max.X
			jp.Height = imgSrc.Bounds().Max.Y
			jp.Type = "image/png"
			jp.Bytes = data
			return jp, nil
		}
	case "image/jpeg":
		imgSrc, err = jpeg.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
		if rowPhoto.Size() < int64(maxSize) {
			jp.Width = imgSrc.Bounds().Max.X
			jp.Height = imgSrc.Bounds().Max.Y
			jp.Type = "image/jpeg"
			jp.Bytes = data
			return jp, nil
		}
	default:
		imgSrc, _, err = image.Decode(rowPhoto)
		if err != nil {
			return nil, err
		}
	}
	newImg := image.NewRGBA(imgSrc.Bounds())
	draw.Draw(newImg, newImg.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	draw.Draw(newImg, newImg.Bounds(), imgSrc, imgSrc.Bounds().Min, draw.Over)
	newPhoto := bytes.NewBuffer(nil)
	for quality := 100; quality > 0; quality -= 5 {
		newPhoto.Reset()
		err = jpeg.Encode(newPhoto, newImg, &jpeg.Options{Quality: quality})
		if err != nil {
			return nil, err
		}
		if newPhoto.Len() <= int(maxSize) {
			break
		}
	}
	if newPhoto.Len() > int(maxSize) {
		return nil, errors.New("unable to compress within the maximum limit")
	}
	jp.Width = newImg.Bounds().Max.X
	jp.Height = newImg.Bounds().Max.Y
	jp.Bytes = newPhoto.Bytes()
	jp.Type = "image/jpeg"
	return jp, nil
}

func GetType(data []byte) (string, error) {
	filetype := http.DetectContentType(data)
	for i := 0; i < len(ext); i++ {
		if strings.Contains(ext[i], filetype[6:]) {
			return filetype, nil
		}
	}
	return "", errors.New("invalid image type")
}

var ext = []string{
	"ase",
	"art",
	"bmp",
	"blp",
	"cd5",
	"cit",
	"cpt",
	"cr2",
	"cut",
	"dds",
	"dib",
	"djvu",
	"egt",
	"exif",
	"gif",
	"gpl",
	"grf",
	"icns",
	"ico",
	"iff",
	"jng",
	"jpeg",
	"jpg",
	"jfif",
	"jp2",
	"jps",
	"lbm",
	"max",
	"miff",
	"mng",
	"msp",
	"nitf",
	"ota",
	"pbm",
	"pc1",
	"pc2",
	"pc3",
	"pcf",
	"pcx",
	"pdn",
	"pgm",
	"PI1",
	"PI2",
	"PI3",
	"pict",
	"pct",
	"pnm",
	"pns",
	"ppm",
	"psb",
	"psd",
	"pdd",
	"psp",
	"px",
	"pxm",
	"pxr",
	"qfx",
	"raw",
	"rle",
	"sct",
	"sgi",
	"rgb",
	"int",
	"bw",
	"tga",
	"tiff",
	"tif",
	"vtf",
	"xbm",
	"xcf",
	"xpm",
	"3dv",
	"amf",
	"ai",
	"awg",
	"cgm",
	"cdr",
	"cmx",
	"dxf",
	"e2d",
	"egt",
	"eps",
	"fs",
	"gbr",
	"odg",
	"svg",
	"stl",
	"vrml",
	"x3d",
	"sxd",
	"v2d",
	"vnd",
	"wmf",
	"emf",
	"art",
	"xar",
	"png",
	"webp",
	"jxr",
	"hdp",
	"wdp",
	"cur",
	"ecw",
	"iff",
	"lbm",
	"liff",
	"nrrd",
	"pam",
	"pcx",
	"pgf",
	"sgi",
	"rgb",
	"rgba",
	"bw",
	"int",
	"inta",
	"sid",
	"ras",
	"sun",
	"tga",
}
