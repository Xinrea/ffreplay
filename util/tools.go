//go:build windows || linux || darwin

package util

import (
	"log"

	"golang.design/x/clipboard"
)

func SetExitMessage(msg string) {
	log.Println(msg)
}

func CurrentOrigin() string {
	return "https://ffreplay.xinrea.cn"
}

func ReadClipboard() string {
	err := clipboard.Init()
	if err != nil {
		return ""
	}

	return string(clipboard.Read(clipboard.FmtText))
}

func Redirect(url string) {
	log.Println(url)
}

func UpdateLocalStorage(key string, value string) {
	log.Println(key, value)
}

func GetLocalStorage(key string) string {
	log.Println(key)

	return ""
}

func RemoveLocalStorage(key string) {
	log.Println(key)
}
