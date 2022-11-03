package handler

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
)

// SuccessColor is a customizable green color printer.
var SuccessColor = color.New(color.FgGreen)

// ErrorColor is a customizable red color printer.
var ErrorColor = color.New(color.FgRed)

// PrintSuccess is a Printf-compatible func for SuccessColor.
var PrintSuccess = SuccessColor.PrintfFunc()

// PrintError is a Printf-compatible func for ErrorColor.
func PrintError(format string, a ...interface{}) {
	_, _ = ErrorColor.Fprintf(color.Error, format, a)
}

// formatError formats a kitsune.proto.v1.Error to a readable Go error.
func formatError(e *v1.Error) error {
	return errors.New(fmt.Sprintf("%s: %s", e.GetType(), e.GetMsg()))
}
