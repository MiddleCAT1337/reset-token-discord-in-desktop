//go:build windows

package console

import "syscall"

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procSetConsoleOut  = kernel32.NewProc("SetConsoleOutputCP")
	procSetConsoleIn   = kernel32.NewProc("SetConsoleCP")
)

func EnableUTF8() {
	const cpUTF8 = 65001
	procSetConsoleOut.Call(cpUTF8)
	procSetConsoleIn.Call(cpUTF8)
}
