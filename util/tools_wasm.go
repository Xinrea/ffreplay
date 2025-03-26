//go:build js && wasm

package util

import "syscall/js"

func SetExitMessage(msg string) {
	js.Global().Set("exitMessage", msg)
}

func CurrentOrigin() string {
	return js.Global().Get("location").Get("origin").String()
}

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

func Redirect(url string) {
	js.Global().Get("location").Set("href", url)
}

func UpdateLocalStorage(key string, value string) {
	// update map in local storage
	js.Global().Get("localStorage").Call("setItem", key, value)
}

func GetLocalStorage(key string) string {
	value := js.Global().Get("localStorage").Call("getItem", key)

	if value.IsUndefined() || value.IsNull() || value.IsNaN() {
		return ""
	}

	return value.String()
}

func RemoveLocalStorage(key string) {
	js.Global().Get("localStorage").Call("removeItem", key)
}
