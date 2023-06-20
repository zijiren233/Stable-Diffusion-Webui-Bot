package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/zijiren233/go-colorlog"
	translater "github.com/zijiren233/google-translater"
)

func Translate(text string) string {
	translated, err := translater.Translate(
		text,
		"en",
		translater.TranslationParams{
			From: "auto",
		},
	)
	if err != nil {
		translated, err = translater.TranslateWithClienID(text, "en", translater.TranslationWithClienIDParams{
			From: "auto",
		})
		if err != nil {
			colorlog.Errorf("Translate err: %v", err)
			return text
		}
	}
	return translated.Text
}

func Round(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

func GetFileName(filename string) string {
	return filepath.Base(filename)
}

func GetFileNamePrefix(filename string) string {
	file := GetFileName(filename)
	return file[:len(file)-len(GetFileNameExt(filename))]
}

func GetFileNameExt(filename string) string {
	return filepath.Ext(filename)
}

func InString(target string, str_array []string) (int, bool) {
	return In(str_array, func(v string) bool {
		return v == target
	})
}

func In[T comparable](slice []T, fun func(v T) bool) (int, bool) {
	for k, v := range slice {
		if fun(v) {
			return k, true
		}
	}
	return -1, false
}

func DeleteSlice[T comparable](a []T, k T) []T {
	for i, val := range a {
		if val == k {
			return append(a[:i], a[i+1:]...)
		}
	}
	return a
}

func TimeFomate(t time.Duration) string {
	return fmt.Sprintf("%dh %dm %ds", int(t.Hours())%24, int(t.Minutes())%60, int(t.Seconds())%60)
}

func PrintStackTrace(err interface{}) string {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "%v\n", err)
	for i := 1; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
	}
	return buf.String()
}

func GetFile(url string) ([]byte, error) {
	r, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Md5(data []byte) string {
	has := md5.Sum(data)
	return hex.EncodeToString(has[:])
}

func TwoDot(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}

func Retry(frequency uint, enableRecover bool, Delay time.Duration, fun func() (bool, error)) error {
	if enableRecover {
		defer func() {
			if i := recover(); i != nil {
				colorlog.Fatal(i)
			}
		}()
	}
	var (
		err error
		try bool
	)
	for {
		if frequency == 0 {
			return err
		}
		if try, err = fun(); try {
			frequency--
			time.Sleep(Delay)
			continue
		} else {
			return err
		}
	}
}

var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

var lenDefaultLetters = len(defaultLetters)

// RandomString returns a random string with a fixed length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = defaultLetters[rand.Intn(lenDefaultLetters)]
	}
	return string(b)
}
