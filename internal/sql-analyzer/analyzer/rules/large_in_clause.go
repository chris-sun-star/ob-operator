package rules

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type LargeInClauseRule struct {
	*obmysql.BaseOBParserVisitor
	diagnoseResults []model.SqlDiagnoseInfo
	maxInElements   int
}

func NewLargeInClauseRule() *LargeInClauseRule {
	return &LargeInClauseRule{
		BaseOBParserVisitor: &obmysql.BaseOBParserVisitor{},
		maxInElements:       200,
	}
}

func (r *LargeInClauseRule) Name() string {
	return "large_in_clause_rule_adjusted"
}

func (r *LargeInClauseRule) Description() string {
	return "Avoid using IN clauses with more than 200 elements."
}

func (r *LargeInClauseRule) Analyze(tree antlr.ParseTree) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	tree.Accept(r)
	return r.diagnoseResults
}

func (r *LargeInClauseRule) VisitPredicate(ctx *obmysql.PredicateContext) interface{} {
	// predicate : bit_expr IN in_expr ...
	if ctx.IN() != nil {
		inExpr := ctx.In_expr()
		if inExpr != nil {
			ie, ok := inExpr.(*obmysql.In_exprContext)
			if ok {
				// in_expr : select_with_parens | LeftParen expr_list RightParen
				if ie.Expr_list() != nil {
					el, ok := ie.Expr_list().(*obmysql.Expr_listContext)
					if ok {
						// expr_list : bit_expr (Comma bit_expr)*
						count := len(el.AllExpr()) // Assuming AllBit_expr returns all elements
						if count > r.maxInElements {
							r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
								RuleName:   r.Name(),
								Level:      "WARN",
								Suggestion: "Consider alternative strategies like breaking the query into smaller chunks or using EXISTS/JOIN clauses.",
								Reason:     fmt.Sprintf("The IN clause contains %d elements, which exceeds the recommended limit of %d and may degrade query performance.", count, r.maxInElements),
							})
						}
					}
				}
			}
		}
	}
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}
