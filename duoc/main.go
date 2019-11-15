package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

var path string

var interfaceGroup InterfaceGroup

type InterfaceGroup struct {
	PackageName  string
	Imports      []string
	InterfaceDoc []string
	StructName   []string
	interfaces   []InterfaceItem
}

type InterfaceItem struct {
	Methods []MethodItem
}

type MethodItem struct {
	MethodDoc  string
	MethodName string
	Params     []string
	Results    []string
}

func init() {
	flag.StringVar(&path, "i", "", "输入文件路径")
}

func main() {
	flag.Parse()

	if path == "" {
		panic("in path is nil")
	}

	fset := token.NewFileSet()

	path, _ := filepath.Abs(path)
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Println(err)
		return
	}

	ast.Walk(new(Visitor), f)

	pathArr := strings.Split(path, "/")
	outName := strings.Split(pathArr[len(pathArr)-1], ".")[0] + "_compiled.go"
	outPath := strings.Join(pathArr[:len(pathArr)-1], "/") + "/" + outName

	//fmt.Println(outPath)
	//fmt.Printf("%+v\n", interfaceGroup)

	var bf bytes.Buffer
	bf.WriteString("// 根据接口生成的客户端和服务端的代码,需要自己实现服务端函数逻辑\n")
	bf.WriteString("package " + interfaceGroup.PackageName)
	bf.WriteString("\n\n")

	bf.WriteString("import (\n\t")
	bf.WriteString(strings.Join(interfaceGroup.Imports, "\n\t"))
	bf.WriteString("\n")
	bf.WriteString(")\n")

	for k, v := range interfaceGroup.interfaces {

		// 生成服务端代码
		bf.WriteString("type " + interfaceGroup.StructName[k] + " struct {\n")
		bf.WriteString("\tc component.IContainer\n")
		bf.WriteString("}\n\n")

		bf.WriteString("func New" + interfaceGroup.StructName[k] + "(c component.IContainer) *" + interfaceGroup.StructName[k] + "{\n")
		bf.WriteString("\treturn &" + interfaceGroup.StructName[k] + "{c:c}\n")
		bf.WriteString("}\n")

		for _, method := range v.Methods {
			bf.WriteString(method.MethodDoc + "\n")
			bf.WriteString("func(a *" + interfaceGroup.StructName[k] + ") " + method.MethodName + "(" + strings.Join(method.Params, ",") + ")" +
				"(" + strings.Join(method.Results, ",") + ") {\n")
			bf.WriteString("\tpanic(" + `"server not implement ` + method.MethodName + `")` + "\n")
			bf.WriteString("}\n")
			bf.WriteString("\n")
		}

		// 生成客户端代码
		bf.WriteString("type " + interfaceGroup.StructName[k] + "Client struct {\n")
		bf.WriteString("\tclient.RPCClient\n")
		bf.WriteString("}\n\n")

		bf.WriteString("func New" + interfaceGroup.StructName[k] + "Client (client client.RPCClient) *" + interfaceGroup.StructName[k] + "Client {\n")
		bf.WriteString("\treturn &" + interfaceGroup.StructName[k] + "Client{RPCClient:client}\n")
		bf.WriteString("}\n")

		for _, method := range v.Methods {
			bf.WriteString(method.MethodDoc + "\n")
			bf.WriteString("func(c *" + interfaceGroup.StructName[k] + "Client) " + method.MethodName + "(" + strings.Join(method.Params, ",") + ")" +
				"(" + strings.Join(method.Results, ",") + ") {\n\t")
			bf.WriteString(`reply, err := c.Call(ctx, "` + interfaceGroup.PackageName + "." + interfaceGroup.StructName[k] + "/" + method.MethodName + `", req)` + "\n\t")
			bf.WriteString("if err != nil {\n\t\t")
			bf.WriteString("return nil, err\n\t")
			bf.WriteString("}\n\t")
			bf.WriteString("m := &" + strings.Split(method.Results[0], "*")[1] + "{}\n\t")
			bf.WriteString("err = c.RPCClient.Decode(reply, m)\n\t")
			bf.WriteString("if err != nil {\n\t\t")
			bf.WriteString("return nil, err\n\t")
			bf.WriteString("}\n\t")
			bf.WriteString("return m, nil\n")
			bf.WriteString("}\n")
			bf.WriteString("\n")
		}
	}

	ioutil.WriteFile(outPath, bf.Bytes(), 0644)

}
