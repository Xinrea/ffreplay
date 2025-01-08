//go:build js && wasm

package util

import (
	"syscall/js"
)

func ReadClipboard() string {
	readResult := make(chan string, 1)

	js.Global().Get("navigator").Get("clipboard").Call("readText").
		Call("then",
			js.FuncOf(func(this js.Value, args []js.Value) interface{} {
				readResult <- args[0].String()
				return nil
			}),
		).Call("catch",
		js.FuncOf(func(this js.Value, args []js.Value) interface{} {
			println("failed to read clipboard: " + args[0].String())
			readResult <- ""
			return nil
		}),
	)

	return <-readResult
}
