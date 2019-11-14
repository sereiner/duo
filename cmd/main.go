package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
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

	fmt.Println(outPath)
	fmt.Printf("%+v\n", interfaceGroup)
}
