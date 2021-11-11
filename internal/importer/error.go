package importer

import (
	"fmt"
	"text/scanner"
)

type Error struct {
	File string
	Pos  scanner.Position
	Err  error
}

func (e *Error) Error() string {
	if e.File != "" {
		e.Pos.Filename = e.File
	}
	return fmt.Sprintf("%s: %v", e.Pos.String(), e.Err)
}
