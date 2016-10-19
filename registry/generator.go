package registry

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Generates the code for register a set of annotations within provided package.
// Parameters:
// - path - the folder where package source files are located;
// - pck - the shoirt package name;
// - outName - name of output source file; if it is empty then <package>_annotations.go will be used
func GenerateRegistry(path, pck, outName string) {
	// iterate all files within the package (path) and collect all found imports/annotations
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	var allAnnotations []AnnotatedEntry
	var allImports []string
	var foundPackageName string
	for _, file := range files {
		fileName := file.Name()
		if strings.HasSuffix(fileName, ".go") && !strings.HasPrefix(fileName, "_") {
			foundAnnotations, foundImports, foundPackage := ParseFile(path, fileName)
			allAnnotations = append(allAnnotations, foundAnnotations...)
			allImports = combinePackages(allImports, foundImports)
			foundPackageName = foundPackage
		}
	}
	if len(allAnnotations) > 0 {
		combinedAnnotations := combineMethodsAndFields(allAnnotations)
		if outName == "" {
			outName = pck + "_annotations.go"
		} else {
			if !strings.HasSuffix(outName, ".go") {
				outName = outName + ".go"
			}
		}
		content := generateRegistry(combinedAnnotations, foundPackageName, allImports)
		f, err := os.Create(filepath.Join(path, outName))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		bufferedWriter := bufio.NewWriter(f)
		saveContent(bufferedWriter, content)
	}
}

// Combines annotated entries related to the same entry.
// Returns array of combined entries
func combineMethodsAndFields(all []AnnotatedEntry) []AnnotatedEntry {
	// group annotations by common struct name
	chains := make(map[string][]AnnotatedEntry)
	for _, a := range all {
		chains[a.Name] = append(chains[a.Name], a)
	}
	// combine chains
	var combinedAnnotations []AnnotatedEntry
	for _, chain := range chains {
		if len(chain) > 0 {
			combined := AnnotatedEntry{chain[0].Type, chain[0].FullPackage, chain[0].Name, AnnotationsData{}}
			for _, a := range chain {
				combined.AnnotationsData.Self =
					append(combined.AnnotationsData.Self, a.AnnotationsData.Self...)
				combined.AnnotationsData.Fields =
					combineMaps(combined.AnnotationsData.Fields, a.AnnotationsData.Fields)
				combined.AnnotationsData.Methods =
					combineMaps(combined.AnnotationsData.Methods, a.AnnotationsData.Methods)
			}
			combinedAnnotations = append(combinedAnnotations, combined)
		}
	}
	return combinedAnnotations
}

// Adds all entries from source map to target map, replacing the ones already exist in target (if any).
// If target map is nil then new empty map is provided as the target.
// Returns target map as the result
func combineMaps(target, source map[string][]AnnotationDoc) map[string][]AnnotationDoc {
	if target == nil {
		target = make(map[string][]AnnotationDoc)
	}
	for k, v := range source {
		target[k] = v
	}
	return target
}

// Gneretates "package" statement
func generateHeader(packageName string) string {
	packageName = packageName[strings.LastIndex(packageName, "/")+1:]
	return "package " + packageName + "\n\n"
}

// Returns alias for import name in form of "a" + <import number in list of all imports>
func genPackageAlias(i int) string {
	return "a" + strconv.Itoa(i+1)
}

// Generates all import statements with corresponding aliases
func generateImports(imports []string) string {
	var b bytes.Buffer
	// generate import of base package
	b.WriteString("import _base \"github.com/SphereSoftware/go-annotations/registry\"\n")
	// generate other imports
	for i, imp := range imports {
		b.WriteString("import ")
		b.WriteString(genPackageAlias(i))
		b.WriteString(" ")
		b.WriteString(strconv.Quote(imp))
		b.WriteString("\n")
	}
	b.WriteString("\n")
	return b.String()
}

// Replaces full packages names to their aliases to make
// generated code more readable
func replaceImports(content string, imports []string) string {
	// make sure that deeper imports will be replaced first
	// it is required for correct processing of nested packages
	sort.Sort(sort.Reverse(sort.StringSlice(imports)))
	for i, imp := range imports {
		content = strings.Replace(content, imp, genPackageAlias(i), -1)
	}
	return content
}

// Iterates through prepared data and produces the source code for registry
func generateRegistry(all []AnnotatedEntry, foundPackage string, foundImports []string) string {
	var b bytes.Buffer
	var allImports []string
	var allValues bytes.Buffer
	for _, a := range all {
		s, imports := GenerateAnnotationValue(&a, foundPackage, foundImports)
		s = "    _base.Map(" + strconv.Quote(a.FullPackage+"."+a.Name) + ",\n" + s + ")\n"
		allValues.WriteString(s)
		allImports = combinePackages(allImports, imports)
	}
	content := replaceImports(allValues.String(), allImports)
	b.WriteString(generateHeader(foundPackage))
	b.WriteString(generateImports(allImports))
	b.WriteString("func init() {\n")
	b.WriteString(content)
	b.WriteString("\n}\n")
	return b.String()
}
