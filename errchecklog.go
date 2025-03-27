package errchecklog

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var (
	ifacePkg  string
	ifaceName string
)

// Analyzer checks that "if err != nil" blocks contain a call to a specified interface.
var Analyzer = &analysis.Analyzer{
	Name: "errchecklog",
	Doc:  "Checks that 'if err != nil' blocks do NOT contain a call to a specified interface",
	Run:  run,
}

func init() {
	Analyzer.Flags.StringVar(&ifacePkg, "ifacepkg", "", "package path of the interface to look for")
	Analyzer.Flags.StringVar(&ifaceName, "ifacename", "", "name of the interface to look for")
}

func run(pass *analysis.Pass) (interface{}, error) {
	if ifacePkg == "" || ifaceName == "" {
		// If not configured, do nothing.
		return nil, nil
	}

	// Retrieve the built-in error interface.
	errorInterface := types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Body == nil {
				continue
			}

			// Only inspect functions named "handle" or "Handle"
			if funcDecl.Name.Name != "handle" && funcDecl.Name.Name != "Handle" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				if ifStmt, ok := n.(*ast.IfStmt); ok {
					if isErrorNotNil(pass, ifStmt.Cond, errorInterface) {
						if !blockContainsInterfaceCall(pass, ifStmt.Body, ifacePkg, ifaceName) {
							pass.Reportf(
								ifStmt.If,
								"if err != nil block does NOT contain a call to %s.%s",
								ifacePkg,
								ifaceName,
							)
						}
					}
				}
				return true
			})
		}
	}

	return nil, nil
}

// isErrorNotNil checks if 'expr' is exactly 'err != nil'
// where 'err' is assignable to the built-in error interface.
func isErrorNotNil(pass *analysis.Pass, expr ast.Expr, errorType *types.Interface) bool {
	bin, ok := expr.(*ast.BinaryExpr)
	if !ok || bin.Op != token.NEQ {
		return false
	}
	return isErrorType(pass, bin.X, errorType) && isNil(bin.Y)
}

func isErrorType(pass *analysis.Pass, e ast.Expr, errorType *types.Interface) bool {
	t := pass.TypesInfo.TypeOf(e)
	if t == nil {
		return false
	}
	return types.AssignableTo(t, errorType)
}

func isNil(e ast.Expr) bool {
	ident, ok := e.(*ast.Ident)
	return ok && ident.Name == "nil"
}

// blockContainsInterfaceCall checks if there's a call to ifacePkgPath.ifaceName in block.
func blockContainsInterfaceCall(pass *analysis.Pass, block *ast.BlockStmt, ifacePkgPath, ifaceName string) bool {
	if block == nil {
		return false
	}
	found := false
	ast.Inspect(block, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := callExpr.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		typ := pass.TypesInfo.TypeOf(sel.X)
		named, ok := typ.(*types.Named)
		if !ok || named.Obj() == nil || named.Obj().Pkg() == nil {
			return true
		}
		if named.Obj().Pkg().Path() == ifacePkgPath && named.Obj().Name() == ifaceName {
			found = true
			return false // stop searching when found
		}
		return true
	})
	return found
}
