package types

import (
	"fmt"
	"strings"
)

const (
	unknownOperationString = "unknown"

	// UCAN Action Constants for DWN operations (reference from ucan_capabilities.go)
	UCANRecordsWrite       = "records-write"
	UCANRecordsDelete      = "records-delete"
	UCANRecordsRead        = "records-read"
	UCANRecordsQuery       = "records-query"
	UCANProtocolsConfigure = "protocols-configure"
	UCANProtocolsQuery     = "protocols-query"
	UCANPermissionsGrant   = "permissions-grant"
	UCANPermissionsRevoke  = "permissions-revoke"
	UCANPermissionsQuery   = "permissions-query"
	UCANDataSync           = "data-sync"
	UCANMessageSync        = "message-sync"
	UCANCreate             = "create"
	UCANRead               = "read"
	UCANUpdate             = "update"
	UCANDelete             = "delete"
	UCANAdmin              = "admin"
	UCANAll                = "*"
)

// DWNOperation represents different operations that can be performed on DWN
type DWNOperation int

const (
	// Record operations
	RecordCreate DWNOperation = iota
	RecordRead
	RecordUpdate
	RecordDelete
	RecordQuery

	// Protocol operations
	ProtocolInstall
	ProtocolQuery
	ProtocolUpdate

	// Permission operations
	PermissionGrant
	PermissionRevoke
	PermissionQuery

	// Sync operations
	DataSync
	MessageSync

	// Administrative operations
	AdminConfig
	AdminReset
)

// String returns the string representation of the operation
func (op DWNOperation) String() string {
	switch op {
	case RecordCreate:
		return "record_create"
	case RecordRead:
		return "record_read"
	case RecordUpdate:
		return "record_update"
	case RecordDelete:
		return "record_delete"
	case RecordQuery:
		return "record_query"
	case ProtocolInstall:
		return "protocol_install"
	case ProtocolQuery:
		return "protocol_query"
	case ProtocolUpdate:
		return "protocol_update"
	case PermissionGrant:
		return "permission_grant"
	case PermissionRevoke:
		return "permission_revoke"
	case PermissionQuery:
		return "permission_query"
	case DataSync:
		return "data_sync"
	case MessageSync:
		return "message_sync"
	case AdminConfig:
		return "admin_config"
	case AdminReset:
		return "admin_reset"
	default:
		return unknownOperationString
	}
}

// RecordOperation represents specific record operations
type RecordOperation int

const (
	RecordOpCreate RecordOperation = iota
	RecordOpRead
	RecordOpUpdate
	RecordOpDelete
	RecordOpList
)

// String returns the string representation
func (op RecordOperation) String() string {
	switch op {
	case RecordOpCreate:
		return "create"
	case RecordOpRead:
		return "read"
	case RecordOpUpdate:
		return "update"
	case RecordOpDelete:
		return "delete"
	case RecordOpList:
		return "list"
	default:
		return unknownOperationString
	}
}

// ProtocolOperation represents specific protocol operations
type ProtocolOperation int

const (
	ProtocolOpInstall ProtocolOperation = iota
	ProtocolOpQuery
	ProtocolOpUpdate
	ProtocolOpDelete
)

// String returns the string representation
func (op ProtocolOperation) String() string {
	switch op {
	case ProtocolOpInstall:
		return "install"
	case ProtocolOpQuery:
		return "query"
	case ProtocolOpUpdate:
		return "update"
	case ProtocolOpDelete:
		return "delete"
	default:
		return unknownOperationString
	}
}

// PermissionRegistry manages DWN-specific UCAN capability mappings
type PermissionRegistry struct {
	operationCapabilities map[DWNOperation][]string
	recordCapabilities    map[RecordOperation][]string
	protocolCapabilities  map[ProtocolOperation][]string
}

// NewDWNPermissionRegistry creates a new permission registry with UCAN-compatible capabilities
func NewDWNPermissionRegistry() PermissionRegistry {
	registry := PermissionRegistry{
		operationCapabilities: make(map[DWNOperation][]string),
		recordCapabilities:    make(map[RecordOperation][]string),
		protocolCapabilities:  make(map[ProtocolOperation][]string),
	}

	// Initialize UCAN-compatible operation capabilities
	registry.operationCapabilities[RecordCreate] = []string{UCANRecordsWrite, UCANCreate}
	registry.operationCapabilities[RecordRead] = []string{UCANRecordsRead, UCANRead}
	registry.operationCapabilities[RecordUpdate] = []string{UCANRecordsWrite, UCANUpdate}
	registry.operationCapabilities[RecordDelete] = []string{UCANRecordsDelete, UCANDelete}
	registry.operationCapabilities[RecordQuery] = []string{UCANRecordsQuery, UCANRead}

	registry.operationCapabilities[ProtocolInstall] = []string{UCANProtocolsConfigure, UCANCreate, UCANAdmin}
	registry.operationCapabilities[ProtocolQuery] = []string{UCANProtocolsQuery, UCANRead}
	registry.operationCapabilities[ProtocolUpdate] = []string{UCANProtocolsConfigure, UCANUpdate, UCANAdmin}

	registry.operationCapabilities[PermissionGrant] = []string{UCANPermissionsGrant, UCANAdmin}
	registry.operationCapabilities[PermissionRevoke] = []string{UCANPermissionsRevoke, UCANAdmin}
	registry.operationCapabilities[PermissionQuery] = []string{UCANPermissionsQuery, UCANRead}

	registry.operationCapabilities[DataSync] = []string{UCANDataSync, UCANRead, UCANUpdate}
	registry.operationCapabilities[MessageSync] = []string{UCANMessageSync, UCANRead}

	registry.operationCapabilities[AdminConfig] = []string{UCANAdmin}
	registry.operationCapabilities[AdminReset] = []string{UCANAdmin, UCANDelete}

	// Initialize UCAN-compatible record operation capabilities
	registry.recordCapabilities[RecordOpCreate] = []string{UCANRecordsWrite, UCANCreate}
	registry.recordCapabilities[RecordOpRead] = []string{UCANRecordsRead, UCANRead}
	registry.recordCapabilities[RecordOpUpdate] = []string{UCANRecordsWrite, UCANUpdate}
	registry.recordCapabilities[RecordOpDelete] = []string{UCANRecordsDelete, UCANDelete}
	registry.recordCapabilities[RecordOpList] = []string{UCANRecordsQuery, UCANRead}

	// Initialize UCAN-compatible protocol operation capabilities
	registry.protocolCapabilities[ProtocolOpInstall] = []string{UCANProtocolsConfigure, UCANCreate}
	registry.protocolCapabilities[ProtocolOpQuery] = []string{UCANProtocolsQuery, UCANRead}
	registry.protocolCapabilities[ProtocolOpUpdate] = []string{UCANProtocolsConfigure, UCANUpdate}
	registry.protocolCapabilities[ProtocolOpDelete] = []string{UCANProtocolsConfigure, UCANDelete}

	return registry
}

// GetRequiredCapabilities returns the required UCAN capabilities for a DWN operation
func (pr *PermissionRegistry) GetRequiredCapabilities(operation DWNOperation) ([]string, error) {
	capabilities, exists := pr.operationCapabilities[operation]
	if !exists {
		return nil, fmt.Errorf("no capabilities defined for operation: %s", operation.String())
	}
	return capabilities, nil
}

// GetRecordCapabilities returns the required capabilities for a record operation
func (pr *PermissionRegistry) GetRecordCapabilities(operation RecordOperation) []string {
	if capabilities, exists := pr.recordCapabilities[operation]; exists {
		return capabilities
	}
	return []string{"read"} // Default to read permission
}

// GetProtocolCapabilities returns the required capabilities for a protocol operation
func (pr *PermissionRegistry) GetProtocolCapabilities(operation ProtocolOperation) []string {
	if capabilities, exists := pr.protocolCapabilities[operation]; exists {
		return capabilities
	}
	return []string{"read"} // Default to read permission
}

// AddOperationCapabilities adds or updates capabilities for an operation
func (pr *PermissionRegistry) AddOperationCapabilities(
	operation DWNOperation,
	capabilities []string,
) {
	pr.operationCapabilities[operation] = capabilities
}

// AddRecordCapabilities adds or updates capabilities for a record operation
func (pr *PermissionRegistry) AddRecordCapabilities(
	operation RecordOperation,
	capabilities []string,
) {
	pr.recordCapabilities[operation] = capabilities
}

// AddProtocolCapabilities adds or updates capabilities for a protocol operation
func (pr *PermissionRegistry) AddProtocolCapabilities(
	operation ProtocolOperation,
	capabilities []string,
) {
	pr.protocolCapabilities[operation] = capabilities
}

// ParseOperationFromString parses a string into a DWNOperation
func ParseOperationFromString(operationStr string) (DWNOperation, error) {
	switch strings.ToLower(operationStr) {
	case "record_create":
		return RecordCreate, nil
	case "record_read":
		return RecordRead, nil
	case "record_update":
		return RecordUpdate, nil
	case "record_delete":
		return RecordDelete, nil
	case "record_query":
		return RecordQuery, nil
	case "protocol_install":
		return ProtocolInstall, nil
	case "protocol_query":
		return ProtocolQuery, nil
	case "protocol_update":
		return ProtocolUpdate, nil
	case "permission_grant":
		return PermissionGrant, nil
	case "permission_revoke":
		return PermissionRevoke, nil
	case "permission_query":
		return PermissionQuery, nil
	case "data_sync":
		return DataSync, nil
	case "message_sync":
		return MessageSync, nil
	case "admin_config":
		return AdminConfig, nil
	case "admin_reset":
		return AdminReset, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", operationStr)
	}
}

// IsAdminOperation checks if the operation requires admin privileges
func IsAdminOperation(operation DWNOperation) bool {
	switch operation {
	case PermissionGrant, PermissionRevoke, AdminConfig, AdminReset:
		return true
	case ProtocolInstall, ProtocolUpdate:
		return true
	default:
		return false
	}
}

// IsWriteOperation checks if the operation requires write privileges
func IsWriteOperation(operation DWNOperation) bool {
	switch operation {
	case RecordCreate, RecordUpdate, RecordDelete:
		return true
	case ProtocolInstall, ProtocolUpdate:
		return true
	case DataSync:
		return true
	default:
		return false
	}
}

// IsReadOperation checks if the operation requires read privileges
func IsReadOperation(operation DWNOperation) bool {
	switch operation {
	case RecordRead, RecordQuery:
		return true
	case ProtocolQuery:
		return true
	case PermissionQuery:
		return true
	case MessageSync:
		return true
	default:
		return false
	}
}
