package types

import (
	"fmt"
	"strings"

	"github.com/sonr-io/crypto/ucan"
)

// UCAN Action Constants are defined in permissions.go

// UCANCapabilityMapper provides conversion between DWN operations and UCAN capabilities
type UCANCapabilityMapper struct{}

// NewUCANCapabilityMapper creates a new capability mapper
func NewUCANCapabilityMapper() *UCANCapabilityMapper {
	return &UCANCapabilityMapper{}
}

// GetUCANCapabilitiesForOperation returns UCAN-specific capabilities for a DWN operation
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForOperation(operation DWNOperation) []string {
	switch operation {
	case RecordCreate:
		return []string{UCANRecordsWrite, UCANCreate}
	case RecordRead:
		return []string{UCANRecordsRead, UCANRead}
	case RecordUpdate:
		return []string{UCANRecordsWrite, UCANUpdate}
	case RecordDelete:
		return []string{UCANRecordsDelete, UCANDelete}
	case RecordQuery:
		return []string{UCANRecordsQuery, UCANRead}

	case ProtocolInstall:
		return []string{UCANProtocolsConfigure, UCANCreate, UCANAdmin}
	case ProtocolQuery:
		return []string{UCANProtocolsQuery, UCANRead}
	case ProtocolUpdate:
		return []string{UCANProtocolsConfigure, UCANUpdate, UCANAdmin}

	case PermissionGrant:
		return []string{UCANPermissionsGrant, UCANAdmin}
	case PermissionRevoke:
		return []string{UCANPermissionsRevoke, UCANAdmin}
	case PermissionQuery:
		return []string{UCANPermissionsQuery, UCANRead}

	case DataSync:
		return []string{UCANDataSync, UCANRead, UCANUpdate}
	case MessageSync:
		return []string{UCANMessageSync, UCANRead}

	case AdminConfig:
		return []string{UCANAdmin}
	case AdminReset:
		return []string{UCANAdmin, UCANDelete}

	default:
		return []string{UCANRead} // Default to read permission
	}
}

// GetUCANCapabilitiesForRecordOperation returns UCAN capabilities for record operations
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForRecordOperation(operation RecordOperation) []string {
	switch operation {
	case RecordOpCreate:
		return []string{UCANRecordsWrite, UCANCreate}
	case RecordOpRead:
		return []string{UCANRecordsRead, UCANRead}
	case RecordOpUpdate:
		return []string{UCANRecordsWrite, UCANUpdate}
	case RecordOpDelete:
		return []string{UCANRecordsDelete, UCANDelete}
	case RecordOpList:
		return []string{UCANRecordsQuery, UCANRead}
	default:
		return []string{UCANRead}
	}
}

// GetUCANCapabilitiesForProtocolOperation returns UCAN capabilities for protocol operations
func (m *UCANCapabilityMapper) GetUCANCapabilitiesForProtocolOperation(operation ProtocolOperation) []string {
	switch operation {
	case ProtocolOpInstall:
		return []string{UCANProtocolsConfigure, UCANCreate}
	case ProtocolOpQuery:
		return []string{UCANProtocolsQuery, UCANRead}
	case ProtocolOpUpdate:
		return []string{UCANProtocolsConfigure, UCANUpdate}
	case ProtocolOpDelete:
		return []string{UCANProtocolsConfigure, UCANDelete}
	default:
		return []string{UCANRead}
	}
}

// CreateDWNResourceURI builds a DWN resource URI for UCAN validation
func (m *UCANCapabilityMapper) CreateDWNResourceURI(target, resourceType, resourceID string) string {
	if resourceID != "" {
		return fmt.Sprintf("dwn://%s/%s/%s", target, resourceType, resourceID)
	}
	return fmt.Sprintf("dwn://%s/%s", target, resourceType)
}

// CreateRecordResourceURI builds a resource URI for a specific record
func (m *UCANCapabilityMapper) CreateRecordResourceURI(target, recordID string) string {
	return m.CreateDWNResourceURI(target, "records", recordID)
}

// CreateProtocolResourceURI builds a resource URI for a protocol
func (m *UCANCapabilityMapper) CreateProtocolResourceURI(target, protocolURI string) string {
	return m.CreateDWNResourceURI(target, "protocols", protocolURI)
}

// CreatePermissionResourceURI builds a resource URI for permissions
func (m *UCANCapabilityMapper) CreatePermissionResourceURI(target string) string {
	return m.CreateDWNResourceURI(target, "permissions", "")
}

// CreateDWNAttenuation creates a UCAN attenuation for DWN operations
func (m *UCANCapabilityMapper) CreateDWNAttenuation(
	actions []string,
	target, recordType, protocol string,
) ucan.Attenuation {
	resourceURI := m.CreateDWNResourceURI(target, "records", "")

	resource := &ucan.DWNResource{
		SimpleResource: ucan.SimpleResource{
			Scheme: "dwn",
			Value:  fmt.Sprintf("records/%s", target),
			URI:    resourceURI,
		},
		RecordType: recordType,
		Protocol:   protocol,
		Owner:      target,
	}

	capability := &ucan.DWNCapability{
		Actions: actions,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// CreateRecordAttenuation creates a UCAN attenuation for specific record operations
func (m *UCANCapabilityMapper) CreateRecordAttenuation(
	actions []string,
	target, recordID, recordType string,
) ucan.Attenuation {
	resourceURI := m.CreateRecordResourceURI(target, recordID)

	resource := &ucan.DWNResource{
		SimpleResource: ucan.SimpleResource{
			Scheme: "dwn",
			Value:  fmt.Sprintf("records/%s", recordID),
			URI:    resourceURI,
		},
		RecordType: recordType,
		Owner:      target,
	}

	capability := &ucan.DWNCapability{
		Actions: actions,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// CreateProtocolAttenuation creates a UCAN attenuation for protocol operations
func (m *UCANCapabilityMapper) CreateProtocolAttenuation(
	actions []string,
	target, protocolURI string,
) ucan.Attenuation {
	resourceURI := m.CreateProtocolResourceURI(target, protocolURI)

	resource := &ucan.DWNResource{
		SimpleResource: ucan.SimpleResource{
			Scheme: "dwn",
			Value:  fmt.Sprintf("protocols/%s", protocolURI),
			URI:    resourceURI,
		},
		Protocol: protocolURI,
		Owner:    target,
	}

	capability := &ucan.DWNCapability{
		Actions: actions,
	}

	return ucan.Attenuation{
		Capability: capability,
		Resource:   resource,
	}
}

// ValidateUCANCapabilities validates that a UCAN capability grants the required DWN actions
func (m *UCANCapabilityMapper) ValidateUCANCapabilities(
	capability ucan.Capability,
	requiredActions []string,
) bool {
	return capability.Grants(requiredActions)
}

// ConvertLegacyCapabilities converts old string-based capabilities to UCAN format
func (m *UCANCapabilityMapper) ConvertLegacyCapabilities(legacyCapabilities []string) []string {
	var ucanCapabilities []string

	for _, legacy := range legacyCapabilities {
		switch strings.ToLower(legacy) {
		case "write", "create":
			ucanCapabilities = append(ucanCapabilities, UCANCreate)
		case "read", "list", "query":
			ucanCapabilities = append(ucanCapabilities, UCANRead)
		case "update":
			ucanCapabilities = append(ucanCapabilities, UCANUpdate)
		case "delete":
			ucanCapabilities = append(ucanCapabilities, UCANDelete)
		case "admin":
			ucanCapabilities = append(ucanCapabilities, UCANAdmin)
		case "sync":
			ucanCapabilities = append(ucanCapabilities, UCANDataSync)
		case "*":
			ucanCapabilities = append(ucanCapabilities, UCANAll)
		default:
			// Pass through unknown capabilities
			ucanCapabilities = append(ucanCapabilities, legacy)
		}
	}

	return ucanCapabilities
}

// IsUCANAction checks if an action string is a valid UCAN action
func IsUCANAction(action string) bool {
	validActions := []string{
		UCANRecordsWrite, UCANRecordsDelete, UCANRecordsRead, UCANRecordsQuery,
		UCANProtocolsConfigure, UCANProtocolsQuery,
		UCANPermissionsGrant, UCANPermissionsRevoke, UCANPermissionsQuery,
		UCANDataSync, UCANMessageSync,
		UCANCreate, UCANRead, UCANUpdate, UCANDelete,
		UCANAdmin, UCANAll,
	}

	for _, validAction := range validActions {
		if action == validAction {
			return true
		}
	}
	return false
}

// GetDWNCapabilityTemplate returns a preconfigured capability template for DWN
func GetDWNCapabilityTemplate() *ucan.CapabilityTemplate {
	return ucan.StandardDWNTemplate()
}

// UCANPermissionRegistry extends the basic permission registry with UCAN capabilities
type UCANPermissionRegistry struct {
	*PermissionRegistry
	mapper *UCANCapabilityMapper
}

// NewUCANPermissionRegistry creates a new UCAN-aware permission registry
func NewUCANPermissionRegistry() *UCANPermissionRegistry {
	return &UCANPermissionRegistry{
		PermissionRegistry: &PermissionRegistry{
			operationCapabilities: make(map[DWNOperation][]string),
			recordCapabilities:    make(map[RecordOperation][]string),
			protocolCapabilities:  make(map[ProtocolOperation][]string),
		},
		mapper: NewUCANCapabilityMapper(),
	}
}

// GetRequiredUCANCapabilities returns UCAN-specific capabilities for a DWN operation
func (r *UCANPermissionRegistry) GetRequiredUCANCapabilities(operation DWNOperation) ([]string, error) {
	capabilities := r.mapper.GetUCANCapabilitiesForOperation(operation)
	if len(capabilities) == 0 {
		return nil, fmt.Errorf("no UCAN capabilities defined for operation: %s", operation.String())
	}
	return capabilities, nil
}

// GetRecordUCANCapabilities returns UCAN capabilities for record operations
func (r *UCANPermissionRegistry) GetRecordUCANCapabilities(operation RecordOperation) []string {
	return r.mapper.GetUCANCapabilitiesForRecordOperation(operation)
}

// GetProtocolUCANCapabilities returns UCAN capabilities for protocol operations
func (r *UCANPermissionRegistry) GetProtocolUCANCapabilities(operation ProtocolOperation) []string {
	return r.mapper.GetUCANCapabilitiesForProtocolOperation(operation)
}

// CreateDWNAttenuation creates a UCAN attenuation for DWN operations
func (r *UCANPermissionRegistry) CreateDWNAttenuation(
	actions []string,
	target, recordType, protocol string,
) ucan.Attenuation {
	return r.mapper.CreateDWNAttenuation(actions, target, recordType, protocol)
}
