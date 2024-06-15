package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"syscall/js"
)

const defaultSrc = `package foo

import (
	"fmt"
	"time"
)

func bar() {
	fmt.Println(time.Now())
}`

func main() {
	// Register a function in Go that can be called from JavaScript
	js.Global().Set("toAst", js.FuncOf(toAst))
	// prevent the Go program from exiting

	// https://github.com/norunners/vue/issues/40

	select {}
	// <-make(chan bool)
}

type ErrorJson struct {
	Error string `json:"error"`
}

//	{
//	  "Doc": null,
//	  "Package": 1,
//	  "Name": {
//	    "NamePos": 9,
//	    "Name": "main",
//	    "Obj": null
//	  },
//	  "Decls": null,
//	  "FileStart": 1,
//	  "FileEnd": 15,
//	  "Scope": {
//	    "Outer": null,
//	    "Objects": {}
//	  },
//	  "Imports": null,
//	  "Unresolved": null,
//	  "Comments": null,
//	  "GoVersion": ""
//	}

type N struct {
	// ast.Node
	Pos token.Pos `json:"pos"`
	End token.Pos `json:"end"`
}

type AstJson struct {
	// ast.File
	// Doc *ast.CommentGroup
	Doc     *ast.CommentGroup `json:"doc"`     // associated documentation; or nil
	Package token.Pos         `json:"package"` // position of "package" keyword
	Name    *ast.Ident        `json:"name"`    // package name
	Decls   []interface{}     `json:"decls"`   // top-level declarations; or nil

	FileStart  token.Pos           `json:"fileStart"`
	FileEnd    token.Pos           `json:"fileEnd"`    // start and end of entire file
	Scope      *ast.Scope          `json:"scope"`      // package scope (this file only)
	Imports    []*ast.ImportSpec   `json:"importSpec"` // imports in this file
	Unresolved []*ast.Ident        `json:"ident"`      // unresolved identifiers in this file
	Comments   []*ast.CommentGroup `json:"comments"`   // list of all comments in the source file
	GoVersion  string              `json:"goVersion"`
}

func toAst(this js.Value, args []js.Value) interface{} {
	fset := token.NewFileSet() // positions are relative to fset

	var src string
	// fmt.Println("args:", args[0])
	// fmt.Println("string:", args[0].String())
	if len(args) > 0 {
		src = args[0].String()
	}

	// Parse src but stop after processing the imports.
	f, err := parser.ParseFile(fset, "", src, parser.Mode(0))

	if err != nil {
		errorJson := ErrorJson{Error: err.Error()}
		serialized, _ := json.Marshal(errorJson)
		return string(serialized)
	}

	// // Print the imports from the file's AST.
	// for _, s := range f.Imports {
	// 	fmt.Println(s.Path.Value)
	// }

	// get "f" ast as json

	var astJson AstJson

	astJson.Doc = f.Doc
	astJson.Package = f.Package
	astJson.Name = f.Name
	// astJson.Decls = f.Decls

	for _, decl := range f.Decls {

		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			fd := ast.FuncDecl{
				Doc:  funcDecl.Doc,
				Recv: funcDecl.Recv,
				Name: &ast.Ident{
					Name:    funcDecl.Name.Name,
					NamePos: funcDecl.Name.NamePos,
					// Obj:     funcDecl.Name.Obj, // ! causes cycle in json
				},
				Type: funcDecl.Type,
				Body: funcDecl.Body,
			}
			astJson.Decls = append(astJson.Decls, fd)
		}

		if genDecl, ok := decl.(*ast.GenDecl); ok {
			gd := ast.GenDecl{
				Doc:    genDecl.Doc,
				TokPos: genDecl.TokPos,
				Tok:    genDecl.Tok,
				Lparen: genDecl.Lparen,
				// specs = composite literal type... iterate it,
				// handle the exact type, and remove circular references
				// Specs:  &ast.Spec{},
				Rparen: genDecl.Rparen,
			}

			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok {
					vs := ast.ValueSpec{
						Doc: valueSpec.Doc,
						// Names: &ast.Ident{},
						Type:    valueSpec.Type,
						Values:  valueSpec.Values,
						Comment: valueSpec.Comment,
					}
					for _, name := range valueSpec.Names {
						n := ast.Ident{
							Name:    name.Name,
							NamePos: name.NamePos,
							// Obj:     name.Obj, // ! causes cycle in json
						}
						vs.Names = append(vs.Names, &n)
					}
					gd.Specs = append(gd.Specs, &vs)
				}

				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					ts := ast.TypeSpec{
						Doc:     typeSpec.Doc,
						Name:    typeSpec.Name,
						Assign:  typeSpec.Assign,
						Type:    typeSpec.Type,
						Comment: typeSpec.Comment,
					}
					gd.Specs = append(gd.Specs, &ts)
				}

				if importSpec, ok := spec.(*ast.ImportSpec); ok {
					is := ast.ImportSpec{
						Doc:     importSpec.Doc,
						Name:    importSpec.Name,
						Path:    importSpec.Path,
						Comment: importSpec.Comment,
					}
					gd.Specs = append(gd.Specs, &is)
				}
			}

			astJson.Decls = append(astJson.Decls, gd)
		}

		// if genDecl, ok := decl.(*ast.GenDecl); ok {
		// 	astJson.Decls = append(astJson.Decls, genDecl)
		// }
	}

	astJson.FileStart = f.Pos()
	astJson.FileEnd = f.End()
	// astJson.Scope = f.Scope
	astJson.Imports = f.Imports
	astJson.Unresolved = f.Unresolved
	astJson.Comments = f.Comments
	astJson.GoVersion = f.GoVersion

	fmt.Println("astJson:", astJson)

	serialized, err := json.Marshal(astJson)
	if err != nil {
		errorJson := ErrorJson{Error: err.Error()}
		serialized, _ := json.Marshal(errorJson)
		return string(serialized)
	}
	return string(serialized)
}

// Golang exec command: stream output to stdout
// AND capture output in variable
// https://stackoverflow.com/a/72809770/9823455
// https://go.dev/play/p/T-o3QvGOm5q
