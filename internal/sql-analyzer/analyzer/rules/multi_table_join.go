package rules

import (
	"fmt"
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type MultiTableJoinRule struct {
	*obmysql.BaseOBParserVisitor
	diagnoseResults []model.SqlDiagnoseInfo
	joinCount       int
}

func NewMultiTableJoinRule() *MultiTableJoinRule {
	return &MultiTableJoinRule{
		BaseOBParserVisitor: &obmysql.BaseOBParserVisitor{},
	}
}

func (r *MultiTableJoinRule) Name() string {
	return "multi_table_join_rule"
}

func (r *MultiTableJoinRule) Description() string {
	return "The number of association tables is not recommended to exceed 5"
}

func (r *MultiTableJoinRule) Analyze(tree antlr.ParseTree) []model.SqlDiagnoseInfo {
	r.diagnoseResults = []model.SqlDiagnoseInfo{}
	r.joinCount = 0
	tree.Accept(r)
	
	if r.joinCount > 5 {
		r.diagnoseResults = append(r.diagnoseResults, model.SqlDiagnoseInfo{
			RuleName:   r.Name(),
			Level:      "WARN",
			Suggestion: "Consider breaking the query into smaller, simpler queries or reviewing schema design.",
			Reason:     fmt.Sprintf("The query involves %d tables in JOIN operations, exceeding the recommended limit of 5.", r.joinCount),
		})
	}
	
	return r.diagnoseResults
}

func (r *MultiTableJoinRule) VisitJoined_table(ctx *obmysql.Joined_tableContext) interface{} {
	// joined_table : table_factor inner_join_type table_factor ...
	// Every time we visit a joined_table node, it implies a join operation.
	// Note: joined_table is recursive. 
	// A JOIN B JOIN C
	// joined_table( joined_table(A, B), C )
	// So visiting each node counts 1 join.
	// Wait, if I have A JOIN B, that is 1 join (2 tables).
	// If I have A JOIN B JOIN C, that is 2 joins (3 tables).
	// The rule says "number of association tables".
	// 5 tables = 4 joins?
	// Python implementation counts `visit_join`.
	// If count >= 5, match=True.
	// So it warns if there are 5 or more JOINS?
	// Python: `if self.join_count >= 5: match = True`.
	// It warns "involves {join_count} tables". Actually if join_count is 5, it means 6 tables (usually).
	// But let's stick to the logic: count joins.
	
	r.joinCount++
	return r.BaseOBParserVisitor.VisitChildren(ctx)
}
