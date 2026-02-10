package executor

import (
	"errors"
	"go/parser"
	"go/token"
	"strings"
)

var (
	ErrForbiddenImport = errors.New("forbidden import: use of os/exec, syscall, or unsafe is not allowed")
)

var forbiddenPackages = map[string]bool{
	"os/exec": true,
	"syscall": true,
	"unsafe": true,
}

// ValidateCode checks that the code does not import forbidden packages.
func ValidateCode(code string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", code, parser.ImportsOnly)
	if err != nil {
		return err
	}
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if forbiddenPackages[path] {
			return ErrForbiddenImport
		}
	}
	return nil
}
