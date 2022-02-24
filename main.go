package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	panicPkg   = "panik"
	panicFn    = "Catch"
	panicDefer = "defer " + panicPkg + "." + panicFn + "()"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Path argument is not defined")
	}

	rootPath := os.Args[1]

	pathInfo, err := os.Lstat(rootPath)
	if err != nil {
		log.Fatalf("Failed to locate path:'%s' error: %v", rootPath, err)
	}

	if !pathInfo.IsDir() {
		log.Fatalf("Path '%s' is not a directory", rootPath)
	}

	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Failed to walk dir path '%s' error: %v", path, err)
			return nil
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".go" {
			return nil
		}

		findGoStmts(path)

		return nil

	})

	if err != nil {
		log.Fatalf("filepath.Walkdir error: %v", err)
	}
}

func findGoStmts(path string) {
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(err)
	}

	// DEBUG
	var v visitor
	ast.Walk(v, astFile)

	ast.Inspect(astFile, func(n ast.Node) bool {
		goStmt, ok := n.(*ast.GoStmt)
		if !ok {
			return true
		}

		fmt.Printf("Go statement found at file: %s position: %d \n", path, goStmt.Pos())

		call := goStmt.Call
		var err error

		switch fun := call.Fun.(type) {
		case *ast.FuncLit:
			if !checkFuncLit(fun) {
				err = fmt.Errorf("first statement should be '%s'", panicDefer)
			}
		case *ast.Ident:
			if fun.Name != panicFn {
				err = fmt.Errorf("deferred function should be '%s()'", panicFn)
			}
		case *ast.SelectorExpr:
			pkg, ok := fun.X.(*ast.Ident)
			if ok {
				if pkg.Name != panicPkg {
					err = fmt.Errorf("deferred function should call '%s()' in '%s' package", panicFn, panicPkg)
				}
			}

			if fun.Sel.Name != panicFn {
				err = fmt.Errorf("deferred function should be '%s()'", panicFn)
			}
		default:
			err = fmt.Errorf("go statement should always call a func lit")
		}

		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			return false
		}

		fmt.Println("Successful statement :)")

		return true
	})
}

func checkFuncLit(fl *ast.FuncLit) bool {
	if len(fl.Body.List) == 0 {
		return true
	}

	firstStmt := fl.Body.List[0]
	deferStmt, ok := firstStmt.(*ast.DeferStmt)
	if !ok {
		return false
	}

	callSel, ok := deferStmt.Call.Fun.(*ast.SelectorExpr)
	if !ok || callSel.Sel.Name != panicFn {
		return false
	}

	return true
}

type visitor int

func (v visitor) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return nil
	}
	fmt.Printf("%s%T\n", strings.Repeat("\t", int(v)), n)
	return v + 1
}
