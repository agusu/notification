package models

import "notification/channels"

// ChannelSchemasResponse represents the response for channel metadata schemas
type ChannelSchemasResponse struct {
	Email channels.ValidEmailMeta `json:"email"`
	SMS   channels.ValidSMSMeta   `json:"sms"`
	Push  channels.ValidPushMeta  `json:"push"`
}
