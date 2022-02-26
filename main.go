package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
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

	findGoStmts(os.Args[1])
}

func findGoStmts(path string) {
	fset := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fset, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("Failed to parse dir %s Error: %v", path, err)
	}

	for _, pkg := range astPkgs {
		for _, astFile := range pkg.Files {

			// DEBUG. Uncomment to see the tree view
			// var v visitor
			// ast.Walk(v, astFile)

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
	}
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
