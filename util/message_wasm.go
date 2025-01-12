//go:build js && wasm

package util

import "syscall/js"

func SetExitMessage(msg string) {
	js.Global().Set("exitMessage", msg)
}
