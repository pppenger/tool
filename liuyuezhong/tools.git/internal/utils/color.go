package utils

import "github.com/fatih/color"

func PrintNormal(format string, a ...interface{}) {
	_, _ = color.New(color.FgGreen).Printf(format+"\n", a...)
}

func PrintInfo(format string, a ...interface{}) {
	_, _ = color.New(color.FgBlue).Printf(format+"\n", a...)
}

func PrintWarn(format string, a ...interface{}) {
	_, _ = color.New(color.FgBlue).Printf(format+"\n", a...)
}

func PrintError(format string, a ...interface{}) {
	_, _ = color.New(color.FgRed).Printf(format+"\n", a...)
}
