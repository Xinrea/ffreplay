//go:build !js

package ui

import (
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/Xinrea/ffreplay/internal/planner"
	"golang.design/x/clipboard"
)

func openedSharedBoard() bool { return false }

func readGameCode() (string, error) {
	if err := clipboard.Init(); err != nil {
		return "", err
	}
	code := clipboard.Read(clipboard.FmtText)
	if len(code) == 0 {
		return "", errors.New("请先复制游戏分享码，也可拖入包含分享码的文本文件")
	}
	return string(code), nil
}
func writeGameCode(code string) error {
	if err := clipboard.Init(); err != nil {
		return err
	}
	if clipboard.Write(clipboard.FmtText, []byte(code)) == nil {
		return errors.New("无法写入剪贴板")
	}
	return nil
}

func guardBoardExit(bool) {} // beforeunload is a browser lifecycle event.

func boardDirectory() (string, error) {
	dir, err := os.UserConfigDir()
	return filepath.Join(dir, "ffreplay"), err
}

func loadBoard() ([]byte, error) {
	dir, err := boardDirectory()
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filepath.Join(dir, "draft.ffplan"))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, planner.MaxFileSize+1))
}

func writeBoard(dir, pattern string, data []byte) (string, error) {
	f, err := os.CreateTemp(dir, pattern)
	if err != nil {
		return "", err
	}
	_, err = f.Write(data)
	closeErr := f.Close()
	if err = errors.Join(err, closeErr); err != nil {
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

func saveBoard(data []byte) (string, error) {
	dir, err := boardDirectory()
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	name, err := writeBoard(dir, ".draft-*", data)
	if err != nil {
		return "", err
	}
	defer os.Remove(name)
	err = os.Rename(name, filepath.Join(dir, "draft.ffplan"))
	return "已保存草稿，下次打开会恢复", err
}

func exportBoard(data []byte) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	name, err := writeBoard(dir, "strategy-*.ffplan", data)
	return "已导出：" + name, err
}

func (b *StrategyBoard) platformControls(x int) {
	// Native plans are shared via Export; no URL action is necessary.
}
