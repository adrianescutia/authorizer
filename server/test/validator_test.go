package test

import (
	"testing"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/utils"
	"github.com/stretchr/testify/assert"
)

func TestIsValidEmail(t *testing.T) {
	validEmail := "lakhan@gmail.com"
	invalidEmail1 := "lakhan"
	invalidEmail2 := "lakhan.me"

	assert.True(t, utils.IsValidEmail(validEmail), "it should be valid email")
	assert.False(t, utils.IsValidEmail(invalidEmail1), "it should be invalid email")
	assert.False(t, utils.IsValidEmail(invalidEmail2), "it should be invalid email")
}

func TestIsValidOrigin(t *testing.T) {
	// don't use portocal(http/https) for ALLOWED_ORIGINS while testing,
	// as we trim them off while running the main function
	envstore.EnvInMemoryStoreObj.UpdateEnvVariable(constants.SliceStoreIdentifier, constants.EnvKeyAllowedOrigins, []string{"localhost:8080", "*.google.com", "*.google.in", "*abc.*"})
	assert.False(t, utils.IsValidOrigin("http://myapp.com"), "it should be invalid origin")
	assert.False(t, utils.IsValidOrigin("http://appgoogle.com"), "it should be invalid origin")
	assert.True(t, utils.IsValidOrigin("http://app.google.com"), "it should be valid origin")
	assert.False(t, utils.IsValidOrigin("http://app.google.ind"), "it should be invalid origin")
	assert.True(t, utils.IsValidOrigin("http://app.google.in"), "it should be valid origin")
	assert.True(t, utils.IsValidOrigin("http://xyx.abc.com"), "it should be valid origin")
	assert.True(t, utils.IsValidOrigin("http://xyx.abc.in"), "it should be valid origin")
	assert.True(t, utils.IsValidOrigin("http://xyxabc.in"), "it should be valid origin")
	assert.True(t, utils.IsValidOrigin("http://localhost:8080"), "it should be valid origin")
	envstore.EnvInMemoryStoreObj.UpdateEnvVariable(constants.SliceStoreIdentifier, constants.EnvKeyAllowedOrigins, []string{"*"})
}

func TestIsValidIdentifier(t *testing.T) {
	assert.False(t, utils.IsValidVerificationIdentifier("test"), "it should be invalid identifier")
	assert.True(t, utils.IsValidVerificationIdentifier(constants.VerificationTypeBasicAuthSignup), "it should be valid identifier")
	assert.True(t, utils.IsValidVerificationIdentifier(constants.VerificationTypeUpdateEmail), "it should be valid identifier")
	assert.True(t, utils.IsValidVerificationIdentifier(constants.VerificationTypeForgotPassword), "it should be valid identifier")
}
