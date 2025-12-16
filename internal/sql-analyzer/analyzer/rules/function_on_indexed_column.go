package rules

import (
	"strings"

	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type FunctionOnIndexedColumnRule struct {
	*obmysql.BaseOBParserVisitor
	diagnoseResults []model.SqlDiagnoseInfo
	indexes         []model.IndexInfo
	inPredicate     bool
	inFunction      bool
}

func NewFunctionOnIndexedColumnRule() *FunctionOnIndexedColumnRule {
	return &FunctionOnIndexedColumnRule{
		BaseOBParserVisitor: &obmysql.BaseOBParserVisitor{},
	}
}

func (r *FunctionOnIndexedColumnRule) Name() string {
	return "function_on_indexed_column_rule"
}

func (r *FunctionOnIndexedColumnRule) Description() string {
	return "Avoid wrapping indexed columns in functions within WHERE/ON clauses, as this prevents index usage (e.g., use 'col >= ...' instead of 'YEAR(col) = ...')."
}

func (r *FunctionOnIndexedColumnRule) Analyze(tree antlr.ParseTree, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.indexes = indexes
	r.inPredicate = false
	r.inFunction = false
	tree.Accept(r)
	return r.diagnoseResults
}

func (r *FunctionOnIndexedColumnRule) VisitPredicate(ctx *obmysql.PredicateContext) interface{} {
	oldPredicate := r.inPredicate
	r.inPredicate = true
	defer func() { r.inPredicate = oldPredicate }()
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}

func (r *FunctionOnIndexedColumnRule) VisitFunc_expr(ctx *obmysql.Func_exprContext) interface{} {
	if !r.inPredicate {
		return r.BaseOBParserVisitor.VisitChildren(ctx)
	}

	oldFunction := r.inFunction
	r.inFunction = true
	defer func() { r.inFunction = oldFunction }()
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}

func (r *FunctionOnIndexedColumnRule) VisitColumn_ref(ctx *obmysql.Column_refContext) interface{} {
	if r.inPredicate && r.inFunction {
		colName := ctx.Column_name().GetText()
		colName = strings.Trim(colName, "`")

		tableName := ""
		if len(ctx.AllRelation_name()) > 0 {
			tableName = ctx.Relation_name(0).GetText()
			tableName = strings.Trim(tableName, "`")
		}

		if r.isColumnIndexed(tableName, colName) {
			r.addResult(colName)
		}
	}
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}

func (r *FunctionOnIndexedColumnRule) isColumnIndexed(tableName, colName string) bool {
	for _, idx := range r.indexes {
		if tableName != "" && !strings.EqualFold(idx.TableName, tableName) {
			continue
		}
		for _, idxCol := range idx.Columns {
			if strings.EqualFold(idxCol, colName) {
				return true
			}
		}
	}
	return false
}

func (r *FunctionOnIndexedColumnRule) addResult(colName string) {
	r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
		RuleName:   r.Name(),
		Level:      "WARN",
		Suggestion: "Indexed column '" + colName + "' is wrapped in a function in a predicate. This prevents index usage. Consider rewriting the query.",
		Reason:     r.Description(),
	})
}
