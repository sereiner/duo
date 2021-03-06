package main

import (
	"fmt"
	"go/ast"
	"go/token"
)

type Visitor struct {
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	switch node := node.(type) {
	case *ast.File:
		interfaceGroup.PackageName = node.Name.Name

	case *ast.GenDecl:
		if node.Tok == token.IMPORT {
			imports := make([]string, len(node.Specs)+2)
			for k, v := range node.Specs {
				imptSpec := v.(*ast.ImportSpec)
				imports[k] = imptSpec.Path.Value
			}
			imports[len(imports)-1] = `"github.com/sereiner/duo/component"`
			imports[len(imports)-2] = `"github.com/sereiner/duo/client"`
			interfaceGroup.Imports = append(interfaceGroup.Imports, imports...)
		}

	case *ast.InterfaceType:
		parserInterface(node)
	case *ast.TypeSpec:
		structName := node.Name.Name[1:]
		interfaceGroup.StructName = append(interfaceGroup.StructName, structName)
	}

	return v
}

func parserInterface(iface *ast.InterfaceType) {
	interfaceItem := InterfaceItem{}
	if iface.Methods != nil || iface.Methods.List != nil {
		methods := make([]MethodItem, len(iface.Methods.List))
		for k, v := range iface.Methods.List {

			ft, ok := v.Type.(*ast.FuncType)
			if !ok {
				fmt.Println("func not ok")
				continue
			}

			method := MethodItem{}
			if v.Doc != nil {
				method.MethodDoc = v.Doc.List[0].Text
			}
			method.MethodName = v.Names[0].Name
			params := make([]string, len(ft.Params.List))
			results := make([]string, len(ft.Results.List))

			for k, v := range ft.Params.List {

				var name string
				var selName string
				var identName string
				name = v.Names[0].Name
				if expr, ok := v.Type.(*ast.StarExpr); ok {

					if selector, ok := expr.X.(*ast.SelectorExpr); ok {

						selName = selector.Sel.Name
						if ident, ok := selector.X.(*ast.Ident); ok {
							identName = ident.Name
						}

					}
				}
				params[k] = name + " *" + identName + "." + selName
			}

			for k, v := range ft.Results.List {
				var name string
				var selName string
				var identName string
				name = v.Names[0].Name
				if expr, ok := v.Type.(*ast.StarExpr); ok {

					if selector, ok := expr.X.(*ast.SelectorExpr); ok {

						selName = selector.Sel.Name

						if ident, ok := selector.X.(*ast.Ident); ok {
							identName = ident.Name
						}

					}

				}

				if ident, ok := v.Type.(*ast.Ident); ok {

					identName = ident.Name
					results[k] = name + " " + identName
					continue
				}

				results[k] = name + " *" + identName + "." + selName
			}

			method.Params = params
			method.Results = results
			methods[k] = method
		}
		interfaceItem.Methods = append(interfaceItem.Methods, methods...)
	}

	interfaceGroup.interfaces = append(interfaceGroup.interfaces, interfaceItem)
}
