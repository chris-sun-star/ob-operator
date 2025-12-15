package rules

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type FullScanRule struct {
	*obmysql.BaseOBParserVisitor
	diagnoseResults []model.SqlDiagnoseInfo
	hasSargablePred bool
}

func NewFullScanRule() *FullScanRule {
	return &FullScanRule{
		BaseOBParserVisitor: &obmysql.BaseOBParserVisitor{},
	}
}

func (r *FullScanRule) Name() string {
	return "full_scan_rule"
}

func (r *FullScanRule) Description() string {
	return "Online query full table scan is not recommended. Exceptions are: very small table, very low frequency, or small result set."
}

func (r *FullScanRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.hasSargablePred = false
	
tree.Accept(r)

	if !r.hasSargablePred {
		r.addResult()
	}

	return r.diagnoseResults
}

func (r *FullScanRule) addResult() {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Detected a potential full table scan which may impact performance. Consider adding indexes, refining WHERE clauses, or restructuring the query to utilize existing indexes.",
		Reason:     r.Description(),
	})
}

func (r *FullScanRule) VisitBool_pri(ctx *obmysql.Bool_priContext) interface{} {
	if ctx.COMP_EQ() != nil || ctx.COMP_GE() != nil || ctx.COMP_GT() != nil || ctx.COMP_LE() != nil || ctx.COMP_LT() != nil {
		r.hasSargablePred = true
	}
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}

func (r *FullScanRule) VisitPredicate(ctx *obmysql.PredicateContext) interface{} {
	if ctx.IN() != nil {
		if ctx.Not() == nil {
			r.hasSargablePred = true
		}
	} else if ctx.BETWEEN() != nil {
		if ctx.Not() == nil {
			r.hasSargablePred = true
		}
	} else if ctx.LIKE() != nil {
		if ctx.Not() == nil {
			// Use index 0 for Simple_expr as it's a list in generated code
			patternCtx := ctx.Simple_expr(0)
			if patternCtx != nil {
				if r.isLeftFuzzy(patternCtx) {
					// Left fuzzy ('%abc') is NOT SARGable.
				} else {
					r.hasSargablePred = true
				}
			}
		}
	}
	
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}

func (r *FullScanRule) isLeftFuzzy(ctx obmysql.ISimple_exprContext) bool {
	if se, ok := ctx.(*obmysql.Simple_exprContext); ok {
		if ec := se.Expr_const(); ec != nil {
			if l := ec.Literal(); l != nil {
				if cs := l.Complex_string_literal(); cs != nil {
					val := cs.GetText()
					val = strings.Trim(val, "'\"")
					if strings.HasPrefix(val, "%" ) {
						return true
					}
				}
				// Removed STRING_VALUE check as it's covered by Complex_string_literal in parser
			}
		}
	}
	return false
}

func (r *FullScanRule) VisitSimple_expr(ctx *obmysql.Simple_exprContext) interface{} {
	if ctx.EXISTS() != nil {
		// EXISTS implies subquery check. 
		// Following Python logic where NOT EXISTS is considered "Range" (Good).
		// But detecting NOT here is tricky without context.
		// For now, if we see EXISTS, we treat it as SARGable to avoid false positives for subqueries 
		// which might use indexes internally. 
		// Improvement: check parent for NOT.
		r.hasSargablePred = true
	}
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}