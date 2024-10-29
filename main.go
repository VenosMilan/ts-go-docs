package main

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const defaultNameOutputFile = "doc.md"
const defaultInputPath = "./"
const defaultOutputPath = "./" + defaultNameOutputFile

func main() {
	inputPath := defaultInputPath
	outputFile := defaultOutputPath

	rootCmd := &cobra.Command{
		Use:   "app",
		Short: "App CLI - generate documentation",
		Long:  "App generate documentation from Golang structs a save into the MD file, for now support only GO structs",
		Run: func(cmd *cobra.Command, args []string) {
			in, _ := cmd.Flags().GetString("inputPath")
			fmt.Println("Input path: " + in)
			inputPath = in
			of, _ := cmd.Flags().GetString("outputFile")
			fmt.Println("Output path: " + of)
			outputFile = of
		},
	}

	rootCmd.Flags().StringP("inputPath", "i", defaultInputPath, "Path for input folder that contains go files... /home/project/...")
	rootCmd.Flags().StringP("outputFile", "o", defaultOutputPath, "Path for output docs MD file ./documentation.md")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(errors.Wrap(err, "Error execute cobra command"))
	}

	files, err := getGoFiles(inputPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "Error get go files"))
	}

	structs := prepareMapOfStructs(files, inputPath)

	if err = createMarkdown("", "", outputFile, structs); err != nil {
		log.Fatal(err)
	} else {
		log.Infof("Doc MD file was created")
	}
}

func prepareMapOfStructs(files []string, path string) map[string]map[string][]Structures {
	structs := make(map[string]map[string][]Structures)

	for _, file := range files {
		pkgName, structList, err := parseGoFile(file, path)
		if err != nil {
			log.Errorf("error parsing file %s: %v\n", file, err)
			continue
		}
		if len(structList) > 0 {
			if structs[file] == nil {
				structs[file] = make(map[string][]Structures)
			}

			structs[file][pkgName] = structList
		}
	}
	return structs
}

func createMarkdown(projectName, description, outputFile string, structs map[string]map[string][]Structures) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return err
	}

	defer func(file *os.File) {
		if err := file.Close(); err != nil {
			log.Errorf("Error close file: %s", err.Error())
		}
	}(file)

	content := ""

	if projectName != "" {
		content = fmt.Sprintf("# Project name: %s\n\n", projectName)
	}

	if description != "" {
		content = fmt.Sprintf("## Description:\n\n%s\n\n", description)
	}

	content = createIndex(content, structs)
	content = createListOfStructs(content, structs)

	_, err = file.WriteString(content)

	return err
}

func createListOfStructs(content string, structs map[string]map[string][]Structures) string {
	content += "\n## Structures\n\n"

	for _, pkgAndStructs := range structs {
		for _, structList := range pkgAndStructs {
			for _, strct := range structList {
				content += fmt.Sprintf("#### %s\n\n", strct.StructName)
				content += fmt.Sprintf("")
				content += prepareBodyOfStruct(&strct)
				content += "\n\n"
			}
		}
	}
	return content
}

func createIndex(content string, structs map[string]map[string][]Structures) string {
	content += "\n## Index: \n\n"
	for file, pkgAndStructs := range structs {
		for pkgName, structList := range pkgAndStructs {
			splitFileName := strings.Split(file, "/")

			content += fmt.Sprintf("### File: %s \n ### Package `%s`\n\n", splitFileName[len(splitFileName)-1], pkgName)

			for _, strct := range structList {
				content += fmt.Sprintf("- [%s](#%s)\n", strct.StructName, strings.ToLower(strct.StructName))
			}

			content += "\n"
		}
	}
	return content
}

func prepareBodyOfStruct(strct *Structures) string {
	maxLenOfField := 0
	for _, field := range strct.StructDetail {
		if len(field.FieldName) > maxLenOfField {
			maxLenOfField = len(field.FieldName)
		}
	}

	maxLenOfType := 0
	for _, field := range strct.StructDetail {
		if len(field.FieldType) > maxLenOfType {
			maxLenOfType = len(field.FieldType)
		}
	}

	content := "```go\n"

	if strct.Comment != "" {
		commentWithoutSuffix, _ := strings.CutSuffix(strct.Comment, "\n")
		strct.Comment = commentWithoutSuffix

		if strings.Contains(strct.Comment, "\n") {
			strct.Comment = strings.ReplaceAll(strct.Comment, "\n", "\n"+`\\`)
		}

		content += fmt.Sprintf(`\\%s`, strct.Comment)
		content += fmt.Sprintf("\n")
	}

	content += fmt.Sprintf("type %s struct {\n", strct.StructName)

	for _, f := range strct.StructDetail {
		paddingField := strings.Repeat(" ", maxLenOfField-len(f.FieldName)+1)
		paddingType := strings.Repeat(" ", maxLenOfType-len(f.FieldType)+1)

		content += fmt.Sprintf("  %s%s%s%s%s  %s\n", f.FieldName, paddingField, f.FieldType, paddingType, f.Tag, f.Comment)
	}

	content += "}\n"
	content += "```"

	return content
}

func getGoFiles(root string) ([]string, error) {
	log.Debugf("Trying get go lang files from root %s", root)
	var goFiles []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".go" {
			goFiles = append(goFiles, path)
		}
		return nil
	})

	log.Debugf("Get files was succefull")
	log.Debugf(strings.Join(goFiles, "\n"))
	return goFiles, err
}

func parseGoFile(filename, path string) (string, []Structures, error) {
	structs := make([]Structures, 0)
	fset := token.NewFileSet()

	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return "", nil, err
	}

	packageName := node.Name.Name

	d, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return "", nil, err
	}

	mapOfComment := make(map[string]string)

	for _, f := range d {
		p := doc.New(f, path, 0)

		for _, t := range p.Types {
			mapOfComment[t.Name] = t.Doc
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if ok {
			if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
				s := Structures{
					StructName:   typeSpec.Name.Name,
					StructDetail: make([]Structure, 0),
				}

				s.Comment = mapOfComment[typeSpec.Name.Name]

				for _, f := range typeSpec.Type.(*ast.StructType).Fields.List {
					fieldType := getTypeString(f.Type)

					for _, name := range f.Names {
						detailOfStruct := Structure{
							FieldName: name.Name,
							FieldType: fieldType,
						}

						if f.Tag != nil {
							detailOfStruct.Tag = f.Tag.Value
						}

						s.StructDetail = append(s.StructDetail, detailOfStruct)
					}
				}

				structs = append(structs, s)
			}
		}
		return true
	})

	return packageName, structs, nil
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		return "[]" + getTypeString(t.Elt)
	case *ast.SelectorExpr:
		return getTypeString(t.X) + "." + t.Sel.Name
	default:
		return fmt.Sprintf("%T", expr)
	}
}

type Structures struct {
	Comment      string
	StructName   string
	StructDetail []Structure
}

type Structure struct {
	FieldName string
	FieldType string
	Comment   string
	Tag       string
}
