//go:build windows || linux || darwin

package util

import "log"

func SetExitMessage(msg string) {
	log.Println(msg)
}
