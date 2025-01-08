//go:build windows || linux || darwin

package util

import "golang.design/x/clipboard"

func ReadClipboard() string {
	err := clipboard.Init()
	if err != nil {
		return ""
	}

	return string(clipboard.Read(clipboard.FmtText))
}
