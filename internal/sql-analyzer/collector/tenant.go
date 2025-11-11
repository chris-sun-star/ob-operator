package collector

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/oceanbase/ob-operator/internal/sql-analyzer/model"
	"github.com/oceanbase/ob-operator/internal/sql-analyzer/oceanbase"
)

// getTenantIDByName queries the cluster for a tenant's ID based on its name.
func getTenantIDByName(ctx context.Context, connMgr *oceanbase.ConnectionManager, tenantName string) (uint64, error) {
	manager, err := connMgr.GetSysReadonlyConnection()
	if err != nil {
		return 0, fmt.Errorf("failed to get connection for tenant ID retrieval: %w", err)
	}
	var tenant model.Tenant
	err = manager.QueryRow(ctx, &tenant, "SELECT tenant_id FROM __all_tenant WHERE tenant_name = ?", tenantName)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("tenant '%s' not found", tenantName)
		}
		return 0, err
	}
	return tenant.ID, nil
}
