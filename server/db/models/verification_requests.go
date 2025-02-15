package models

import "github.com/authorizerdev/authorizer/server/graph/model"

// VerificationRequest model for db
type VerificationRequest struct {
	Key        string `json:"_key,omitempty" bson:"_key"` // for arangodb
	ID         string `gorm:"primaryKey;type:char(36)" json:"_id" bson:"_id"`
	Token      string `gorm:"type:text" json:"token" bson:"token"`
	Identifier string `gorm:"uniqueIndex:idx_email_identifier" json:"identifier" bson:"identifier"`
	ExpiresAt  int64  `json:"expires_at" bson:"expires_at"`
	CreatedAt  int64  `gorm:"autoCreateTime" json:"created_at" bson:"created_at"`
	UpdatedAt  int64  `gorm:"autoUpdateTime" json:"updated_at" bson:"updated_at"`
	Email      string `gorm:"uniqueIndex:idx_email_identifier" json:"email" bson:"email"`
}

func (v *VerificationRequest) AsAPIVerificationRequest() *model.VerificationRequest {
	return &model.VerificationRequest{
		ID:         v.ID,
		Token:      &v.Token,
		Identifier: &v.Identifier,
		Expires:    &v.ExpiresAt,
		CreatedAt:  &v.CreatedAt,
		UpdatedAt:  &v.UpdatedAt,
		Email:      &v.Email,
	}
}
