package constants

const (
	AppName = "Kronos"

	IDField = "id"

	AttributesField = "attributes"

	CreatedByField  = "created_by"
	ModifiedByField = "modified_by"

	ModifiedByDBusAPIName = "DBUS_API"
	ModifiedByHTTPAPIName = "HTTP_API"
	ModifiedBySyncName    = "SYNC"

	//OperationTransaction = "TRANSACTION"
	SyncPolicyDontSync = "DONT_SYNC"
)

var (
	// MetaFields is a list of all meta columns of DB entities
	MetaFields = []string{
		"version",
		"sync_version",
		"created_at",
		"modified_at",
		"deleted_at",
	}
)
