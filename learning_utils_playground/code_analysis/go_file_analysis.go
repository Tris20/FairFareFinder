package code_analysis

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

func GetFunctionsInDir(root string) {
	err := filepath.Walk(root, visitFile)
	if err != nil {
		fmt.Println("Error walking through files:", err)
	}
}

func visitFile(path string, fi os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}
	if !fi.IsDir() && filepath.Ext(path) == ".go" {
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			fmt.Println("Error parsing file:", err)
			return err
		}

		ast.Inspect(node, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}
			params := []string{}
			for _, p := range fn.Type.Params.List {
				for _, n := range p.Names {
					params = append(params, fmt.Sprintf("%v %v", n, p.Type))
				}
			}
			fmt.Printf("Function: %s, Parameters: %s\n", fn.Name.Name, params)
			return false
		})
	}
	return nil
}
