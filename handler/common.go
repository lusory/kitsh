package handler

import (
	"errors"
	"fmt"
	"github.com/lusory/libkitsune/proto/kitsune/proto/v1"
)

// FormatError formats a kitsune.proto.v1.Error to a readable Go error.
func FormatError(e *v1.Error) error {
	return errors.New(fmt.Sprintf("%s: %s", e.GetType(), e.GetMsg()))
}
