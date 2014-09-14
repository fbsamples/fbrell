// Package pkgpath provides support for directory and file flags where
// the defaults are defined in terms of a standard Go import
// path. This allows for a development environment to default to using
// the GOPATH etc to define the path, but allows for easy overriding
// for production environments.
package pkgpath

import (
	"flag"
	"go/build"
)

type dir struct {
	dest *string
}

func DirVar(dest *string, name string, value string, usage string) {
	*dest = importPath(value)
	flag.Var(&dir{dest: dest}, name, usage)
}

func Dir(name string, value string, usage string) *string {
	dest := new(string)
	DirVar(dest, name, value, usage)
	return dest
}

func (d *dir) String() string {
	return *(d.dest)
}

func (d *dir) Set(val string) error {
	*d.dest = val
	return nil
}

// Find the package relative directory to use as the browserify directory.
func importPath(path string) string {
	// failures are ignored by design and imply an empty default value.
	pkg, err := build.Import(path, "", build.FindOnly)
	if err != nil {
		return ""
	}
	return pkg.Dir
}
