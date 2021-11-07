package importer

import (
	"fmt"
	"text/scanner"
)

type Error struct {
	Pos scanner.Position
	Err error
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d:%d: %v", e.Pos.Line, e.Pos.Column, e.Err)
}
