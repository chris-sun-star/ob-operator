package oceanbase

const (
	BootstrapTimeoutSeconds = 300
)

const (
	SqlPort = 2881
	RpcPort = 2882
)

const (
	SqlPortName = "sql"
	RpcPortName = "rpc"
)

const (
	ProbeCheckPeriodSeconds = 2
	ProbeCheckDelaySeconds  = 5
)

const (
	ContainerName      = "observer"
	InstallPath        = "/home/admin/oceanbase"
	DataPath           = "/home/admin/data-file"
	ClogPath           = "/home/admin/data-log"
	LogPath            = "/home/admin/log"
	BackupPath         = "/ob-backup"
	DataVolumeSuffix   = "data-file"
	ClogVolumeSuffix   = "data-log"
	LogVolumeSuffix    = "log"
	BackupVolumeSuffix = "backup"
)

const (
	RootUser     = "root"
	ProxyUser    = "proxyro"
	OperatorUser = "operator"
)
