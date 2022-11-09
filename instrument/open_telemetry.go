package instrument

import (
	"go/ast"
	"go/token"
)

type OpenTelemetry struct {
	TracerName  string
	ContextName string
	ErrorName   string
	setError    bool
}

func (s *OpenTelemetry) Imports() []string {
	pkg := []string{
		"go.opentelemetry.io/otel",
	}
	if s.setError {
		pkg = append(pkg, "go.opentelemetry.io/otel/codes")
	}
	return pkg
}

func (s *OpenTelemetry) PrefixStatements(spanName string, hasError bool) []ast.Stmt {
	if hasError {
		s.setError = hasError
	}

	stmts := []ast.Stmt{
		&ast.AssignStmt{
			Tok: token.DEFINE,
			Lhs: []ast.Expr{&ast.Ident{Name: s.ContextName}, &ast.Ident{Name: "span"}},
			Rhs: []ast.Expr{s.expFuncSet(s.TracerName, spanName)},
		},
		&ast.DeferStmt{Call: &ast.CallExpr{
			Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "span"}, Sel: &ast.Ident{Name: "End"}},
		}},
	}
	if hasError {
		stmts = append(stmts, &ast.DeferStmt{Call: &ast.CallExpr{Fun: s.exprFuncSetSpanError(s.ErrorName)}})
	}
	return stmts
}

func (s *OpenTelemetry) expFuncSet(tracerName, spanName string) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X: &ast.CallExpr{
				Fun:  &ast.SelectorExpr{X: &ast.Ident{Name: "otel"}, Sel: &ast.Ident{Name: "Trace"}},
				Args: []ast.Expr{&ast.BasicLit{Kind: token.STRING, Value: `"` + tracerName + `"`}},
			},
			Sel: &ast.Ident{Name: "Start"},
		},
		Args: []ast.Expr{&ast.Ident{Name: "ctx"}, &ast.BasicLit{Kind: token.STRING, Value: `"` + spanName + `"`}},
	}
}

func (s *OpenTelemetry) exprFuncSetSpanError(errorName string) ast.Expr {
	return &ast.FuncLit{
		Type: &ast.FuncType{},
		Body: &ast.BlockStmt{List: []ast.Stmt{
			&ast.IfStmt{
				Cond: &ast.BinaryExpr{X: &ast.Ident{Name: "err"}, Op: token.NEQ, Y: &ast.Ident{Name: "nil"}},
				Body: &ast.BlockStmt{List: []ast.Stmt{
					&ast.ExprStmt{X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "span"}, Sel: &ast.Ident{Name: "SetStatus"}},
						Args: []ast.Expr{
							&ast.SelectorExpr{X: &ast.Ident{Name: "codes"}, Sel: &ast.Ident{Name: "Error"}},
							&ast.BasicLit{Kind: token.STRING, Value: `"error"`},
						},
					}},
					&ast.ExprStmt{X: &ast.CallExpr{
						Fun: &ast.SelectorExpr{X: &ast.Ident{Name: "span"}, Sel: &ast.Ident{Name: "RecordError"}},
						Args: []ast.Expr{
							&ast.Ident{Name: errorName},
						},
					}},
				}},
			},
		}},
	}
}