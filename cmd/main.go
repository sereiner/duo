package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var bf bytes.Buffer

var structName string

func main() {
	fset := token.NewFileSet()
	// 这里取绝对路径，方便打印出来的语法树可以转跳到编辑器
	path, _ := filepath.Abs("/Users/wule/lib/duo/_test/pb/user.go")
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return
	}

	// 打印语法树
	//ast.Print(fset, f)

	ast.Walk(&Visitor{}, f)

	//fmt.Println(bf.String())

	ioutil.WriteFile("./my.go", bf.Bytes(), 0644)

}

// Visitor
type Visitor struct {
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {
	switch node.(type) {
	case *ast.File:
		packName := node.(*ast.File)
		bf.WriteString("package " + packName.Name.Name + "\n")
		bf.WriteString("\n")
	case *ast.Comment:
		comment := node.(*ast.Comment)

		bf.WriteString("\n")
		bf.WriteString(comment.Text + "\n")
	case *ast.GenDecl:
		genDecl := node.(*ast.GenDecl)
		// 查找有没有import context包
		// Notice：没有考虑没有import任何包的情况
		if genDecl.Tok == token.IMPORT {
			v.addImport(genDecl)
			// 不需要再遍历子树
			return nil
		}

	case *ast.InterfaceType:
		// 遍历所有的接口类型
		iface := node.(*ast.InterfaceType)

		addContext(iface)
		// 不需要再遍历子树
		return nil
	case *ast.TypeSpec:
		typeSpec := node.(*ast.TypeSpec)

		structName = typeSpec.Name.Name[1:]
		bf.WriteString(`type ` + structName + ` struct {`)
		bf.WriteString("\n")
		bf.WriteString(`	c component.IContainer `)
		bf.WriteString("\n")
		bf.WriteString("}")
		bf.WriteString("\n")
	}
	return v
}

// addImport 引入context包
func (v *Visitor) addImport(genDecl *ast.GenDecl) {

	imports := make([]string, len(genDecl.Specs)+1)
	for k, v := range genDecl.Specs {
		imptSpec := v.(*ast.ImportSpec)
		imports[k] = imptSpec.Path.Value
	}
	imports = append(imports, `"github.com/sereiner/duo/component"`)

	bf.WriteString("import (\n")
	bf.WriteString(strings.Join(imports, "\n"))
	bf.WriteString("\n")
	bf.WriteString(")\n")

}

// addContext 添加context参数
func addContext(iface *ast.InterfaceType) {
	// 接口方法不为空时，遍历接口方法
	if iface.Methods != nil || iface.Methods.List != nil {

		for _, v := range iface.Methods.List {
			ft, ok := v.Type.(*ast.FuncType)
			if !ok {
				fmt.Println("func not ok")
				continue
			}
			bf.WriteString("\n")
			bf.WriteString(v.Doc.List[0].Text)
			bf.WriteString("\n")

			params := make([]string, len(ft.Params.List))
			results := make([]string, len(ft.Results.List))
			//hasContext := false
			// 判断参数中是否包含context.Context类型
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

			bf.WriteString("func(a *" + structName + ") " + v.Names[0].Name + "(" + strings.Join(params, ",") + ") (" + strings.Join(results, ",") + ") {\n")
			bf.WriteString(" panic(" + `"not implement"` + ")  \n")
			bf.WriteString("}\n")

		}
	}
}
