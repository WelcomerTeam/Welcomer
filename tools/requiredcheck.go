package main

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/unitchecker"
)

var Analyzer = &analysis.Analyzer{
	Name: "requiredcheck",
	Doc:  "checks for missing required struct fields",
	Run:  run,
}

func main() {
	unitchecker.Main(Analyzer)
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		commentMap := ast.NewCommentMap(pass.Fset, file, file.Comments)

		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok {
				return true
			}

			// Check if whole struct literal is ignored
			if hasNoLint(commentMap, lit) {
				return true
			}

			// Get ignored fields for this node
			ignoredFields := getIgnoredFields(commentMap, lit)

			typ := pass.TypesInfo.TypeOf(lit.Type)
			if typ == nil {
				return true
			}

			underlying := typ.Underlying()

			structType, ok := underlying.(*types.Struct)
			if !ok {
				return true
			}

			named, ok := typ.(*types.Named)
			if !ok {
				return true
			}

			// Ignore external packages
			if !strings.Contains(named.Obj().Pkg().Path(), "github.com/WelcomerTeam/Welcomer") {
				// println("Skipping external package:", named.Obj().Pkg().Path())
				return true
			}

			// Track provided fields
			provided := map[string]bool{}
			for _, elt := range lit.Elts {
				if kv, ok := elt.(*ast.KeyValueExpr); ok {
					if ident, ok := kv.Key.(*ast.Ident); ok {
						provided[ident.Name] = true
					}
				}
			}

			// Check required fields
			for i := 0; i < structType.NumFields(); i++ {
				field := structType.Field(i)

				// Skip ignored fields
				if ignoredFields[field.Name()] {
					continue
				}

				// Skip ignored fields via struct tag
				tag := structType.Tag(i)
				if strings.Contains(tag, "requiredcheck:\"ignore\"") {
					continue
				}

				if !provided[field.Name()] {
					pass.Reportf(lit.Pos(), "missing required field: %s (%s)", field.Name(), named.Obj().Pkg().Path())
				}
			}

			return true
		})
	}
	return nil, nil
}

func hasNoLint(cm ast.CommentMap, node ast.Node) bool {
	groups := cm[node]
	for _, group := range groups {
		for _, c := range group.List {
			if strings.Contains(c.Text, "nolint:requiredcheck") {
				return true
			}
		}
	}
	return false
}

func getIgnoredFields(cm ast.CommentMap, node ast.Node) map[string]bool {
	result := make(map[string]bool)

	groups := cm[node]
	for _, group := range groups {
		for _, c := range group.List {
			text := c.Text

			if strings.Contains(text, "requiredcheck:ignore-fields") {
				parts := strings.Split(text, "requiredcheck:ignore-fields")
				if len(parts) < 2 {
					continue
				}

				fieldList := strings.TrimSpace(parts[1])
				fieldList = strings.TrimPrefix(fieldList, ":")
				fields := strings.Split(fieldList, ",")

				for _, f := range fields {
					name := strings.TrimSpace(f)
					if name != "" {
						result[name] = true
					}
				}
			}
		}
	}

	return result
}
