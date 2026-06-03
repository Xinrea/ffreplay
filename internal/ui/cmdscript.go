package ui

import (
	"os"
	"strings"

	"github.com/Xinrea/ffreplay/internal/system/script"
)

// scriptHandler handles /script commands.
// Usage:
//
//	/script run <filename>  - Run a Lua script file
//	/script exec <code>     - Execute inline Lua code
func (c *CommandHandler) scriptHandler(cmds []string) {
	if len(cmds) == 0 {
		c.AddError("用法: /script run <filename> 或 /script exec <code>")

		return
	}

	runner := script.ActiveRunner
	if runner == nil {
		c.AddError("脚本引擎未初始化")

		return
	}

	switch cmds[0] {
	case "run":
		c.scriptRun(runner, cmds[1:])
	case "exec":
		c.scriptExec(runner, cmds[1:])
	default:
		c.AddError("用法: /script run <filename> 或 /script exec <code>")
	}
}

// scriptRun runs a Lua script file.
func (c *CommandHandler) scriptRun(runner *script.ScriptRunner, cmds []string) {
	if len(cmds) == 0 {
		c.AddError("用法: /script run <filename>")

		return
	}

	filename := cmds[0]

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		c.AddError("文件不存在: " + filename)

		return
	}

	c.AddResult("正在运行脚本: " + filename)
	runner.RunFile(filename)
}

// scriptExec executes inline Lua code.
func (c *CommandHandler) scriptExec(runner *script.ScriptRunner, cmds []string) {
	if len(cmds) == 0 {
		c.AddError("用法: /script exec <code>")

		return
	}

	code := strings.Join(cmds, " ")
	c.AddResult("正在执行脚本...")
	runner.Run(code)
}
