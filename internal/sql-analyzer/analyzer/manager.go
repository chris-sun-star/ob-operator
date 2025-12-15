package analyzer

import (
	"github.com/antlr4-go/antlr/v4"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/analyzer/rules"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/api/model"
	// Import the generated parser package.
	// Note: This package path must match where 'make generate-parser' outputs the code.
	obmysql "github.com/oceanbase/ob-operator/internal/sql-analyzer/parser/mysql"
)

type Manager struct {
	rules []Rule
}

func NewManager() *Manager {
	m := &Manager{
		rules: []Rule{},
	}
	m.RegisterRules()
	return m
}

func (m *Manager) RegisterRules() {
	// Register all available rules here
	m.rules = append(m.rules, rules.NewSelectAllRule())
	m.rules = append(m.rules, rules.NewArithmeticRule())
	m.rules = append(m.rules, rules.NewIsNullRule())
	m.rules = append(m.rules, rules.NewLargeInClauseRule())
	m.rules = append(m.rules, rules.NewMultiTableJoinRule())
	m.rules = append(m.rules, rules.NewUpdateDeleteWithoutWhereRule())
	m.rules = append(m.rules, rules.NewUpdateDeleteMultiTableRule())
	m.rules = append(m.rules, rules.NewFullScanRule())
	m.rules = append(m.rules, rules.NewIndexColumnFuzzyMatchRule())
	m.rules = append(m.rules, rules.NewFunctionOnIndexedColumnRule())
}

func (m *Manager) Analyze(sql string, indexes []model.IndexInfo) []model.SqlDiagnoseInfo {
	var diagnostics []model.SqlDiagnoseInfo

	// Setup ANTLR input stream
	inputStream := antlr.NewInputStream(sql)

	// Create Lexer
	lexer := obmysql.NewOBLexer(inputStream)
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)

	// Create Parser
	p := obmysql.NewOBParser(stream)
	// Add error listener to avoid printing to stdout
	p.RemoveErrorListeners()

	// Parse the SQL (assuming 'Sql_stmt' is the entry point rule)
	tree := p.Sql_stmt()

	// Run all registered rules
	for _, rule := range m.rules {
		results := rule.Analyze(tree, indexes)
		diagnostics = append(diagnostics, results...)
	}

	return diagnostics
}