package analyzer

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
)

// Rule is the interface that all SQL review rules must implement.
type Rule interface {
	Name() string
	Description() string
	// Analyze analyzes the parsed SQL tree and returns a list of diagnosis results.
	Analyze(tree antlr.ParseTree) []model.SqlDiagnoseInfo
}
