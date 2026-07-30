package adminlogic

import (
	"wklive/proto/system"
	"wklive/services/system/models"
)

func adminNotificationIncidentToProto(
	value *models.SysAdminNotificationIncident,
) *system.AdminNotificationIncident {
	if value == nil {
		return nil
	}
	return &system.AdminNotificationIncident{
		Id:                   value.Id,
		TenantId:             value.TenantId,
		EventType:            value.EventType,
		AlertKey:             value.AlertKey,
		LastEventId:          value.LastEventId,
		Severity:             value.Severity,
		Title:                value.Title,
		Message:              value.Message,
		Source:               value.Source,
		PayloadJson:          value.PayloadJson,
		Status:               value.Status,
		FirstSeenAt:          value.FirstSeenAt,
		LastSeenAt:           value.LastSeenAt,
		AcknowledgedAt:       value.AcknowledgedAt,
		AcknowledgedBy:       value.AcknowledgedBy,
		AcknowledgedUsername: value.AcknowledgedUsername,
		AcknowledgeReason:    value.AcknowledgeReason,
		EscalationLevel:      value.EscalationLevel,
		LastEscalatedAt:      value.LastEscalatedAt,
		NextEscalateAt:       value.NextEscalateAt,
		ResolvedAt:           value.ResolvedAt,
		Version:              value.Version,
	}
}
