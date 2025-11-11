package model

// PlanIdentifier holds the identifiers for a plan.
type SqlPlanIdentifier struct {
	TenantID uint64
	SvrIP    string
	SvrPort  int64
	PlanID   int64
}

type SqlPlan struct {
	TenantID          uint64 `db:"TENANT_ID"`
	SvrIP             string `db:"SVR_IP"`
	SvrPort           int64  `db:"SVR_PORT"`
	PlanID            int64  `db:"PLAN_ID"`
	SqlID             string `db:"SQL_ID"`
	DbID              int64  `db:"DB_ID"`
	PlanHash          uint64 `db:"PLAN_HASH"`
	GmtCreate         string `db:"GMT_CREATE"`
	Operator          string `db:"OPERATOR"`
	ObjectNode        string `db:"OBJECT_NODE"`
	ObjectID          int64  `db:"OBJECT_ID"`
	ObjectOwner       string `db:"OBJECT_OWNER"`
	ObjectName        string `db:"OBJECT_NAME"`
	ObjectAlias       string `db:"OBJECT_ALIAS"`
	ObjectType        string `db:"OBJECT_TYPE"`
	Optimizer         string `db:"OPTIMIZER"`
	ID                int64  `db:"ID"`
	ParentID          int64  `db:"PARENT_ID"`
	Depth             int64  `db:"DEPTH"`
	Position          int64  `db:"POSITION"`
	Cost              int64  `db:"COST"`
	RealCost          int64  `db:"REAL_COST"`
	Cardinality       int64  `db:"CARDINALITY"`
	RealCardinality   int64  `db:"REAL_CARDINALITY"`
	IoCost            int64  `db:"IO_COST"`
	CpuCost           int64  `db:"CPU_COST"`
	Bytes             int64  `db:"BYTES"`
	Rowset            int64  `db:"ROWSET"`
	OtherTag          string `db:"OTHER_TAG"`
	PartitionStart    string `db:"PARTITION_START"`
	Other             string `db:"OTHER"`
	AccessPredicates  string `db:"ACCESS_PREDICATES"`
	FilterPredicates  string `db:"FILTER_PREDICATES"`
	StartupPredicates string `db:"STARTUP_PREDICATES"`
	Projection        string `db:"PROJECTION"`
	SpecialPredicates string `db:"SPECIAL_PREDICATES"`
	QblockName        string `db:"QBLOCK_NAME"`
	Remarks           string `db:"REMARKS"`
	OtherXML          string `db:"OTHER_XML"`
}
