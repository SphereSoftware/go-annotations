package registry

import (
	"bufio"
	"go/ast"
	"os"
	"path/filepath"
	"strings"
)

var (
	ROOTS = filepath.SplitList(os.Getenv("GOPATH"))
)

// Combines two sets of packages names keeping only unique names.
// Returns the combined set of names.
func combinePackages(allPackages []string, foundPackages []string) []string {
	var combinedPackages []string
	m := make(map[string]bool)
	for _, pck := range allPackages {
		m[pck] = true
		combinedPackages = append(combinedPackages, pck)
	}
	for _, pck := range foundPackages {
		_, found := m[pck]
		if !found {
			combinedPackages = append(combinedPackages, pck)
		}
	}
	return combinedPackages
}

// Finds the folder in file system which contains given package
// It searchs amoung all roots from GOPATH environment variable
// and check where given package exists and returns it back
// Parameter:
// - pck - full package name
func findDirs(pck string) []string {
	var foundPath []string
	for _, root := range ROOTS {
		dir := filepath.Join(root, "src", pck)
		finfo, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			} else {
				panic(err)
			}
		} else if finfo.IsDir() {
			foundPath = append(foundPath, dir)
		}
	}
	return foundPath
}

// Saves provided text content to buffered writer
// Parameters:
// - pointer to buffered writer
// - the string content
func saveContent(out *bufio.Writer, content string) {
	_, err := out.WriteString(content)
	if err == nil {
		err = out.Flush()
		if err == nil {
			return
		}
	}
	panic(err)
}

// Returns full package name by its path
// Parameters:
// - the full path of the package folder
// - short package name (for logging purposes)
func resolveFullPackage(path, shortPackage string) string {
	for _, root := range ROOTS {
		if strings.HasPrefix(path, root) {
			pck := strings.Replace(strings.TrimPrefix(path, root), "\\", "/", -1)
			pck = strings.TrimPrefix(pck, "/src/")
			return pck
		}
	}
	panic("Can't resolve current package '" + shortPackage + "' at path '" + path + "'")
}

// Returns field name from its *ast.Field representation
func getFieldName(f *ast.Field) string {
	if len(f.Names) == 0 {
		panic("Unnamed fields are not supported in annotations")
	}
	if len(f.Names) > 1 {
		panic("Multiple field names found")
	}
	return f.Names[0].Name
}
