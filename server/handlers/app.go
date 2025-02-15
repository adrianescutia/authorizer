package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/utils"
	"github.com/gin-gonic/gin"
)

// State is the struct that holds authorizer url and redirect url
// They are provided via query string in the request
type State struct {
	AuthorizerURL string `json:"authorizerURL"`
	RedirectURL   string `json:"redirectURL"`
}

// AppHandler is the handler for the /app route
func AppHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableLoginPage) {
			c.JSON(400, gin.H{"error": "login page is not enabled"})
			return
		}

		state := c.Query("state")

		var stateObj State

		if state == "" {
			stateObj.AuthorizerURL = envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyAuthorizerURL)
			stateObj.RedirectURL = stateObj.AuthorizerURL + "/app"

		} else {
			decodedState, err := utils.DecryptB64(state)
			if err != nil {
				c.JSON(400, gin.H{"error": "[unable to decode state] invalid state"})
				return
			}

			err = json.Unmarshal([]byte(decodedState), &stateObj)
			if err != nil {
				c.JSON(400, gin.H{"error": "[unable to parse state] invalid state"})
				return
			}
			stateObj.AuthorizerURL = strings.TrimSuffix(stateObj.AuthorizerURL, "/")
			stateObj.RedirectURL = strings.TrimSuffix(stateObj.RedirectURL, "/")

			// validate redirect url with allowed origins
			if !utils.IsValidOrigin(stateObj.RedirectURL) {
				c.JSON(400, gin.H{"error": "invalid redirect url"})
				return
			}

			if stateObj.AuthorizerURL == "" {
				c.JSON(400, gin.H{"error": "invalid authorizer url"})
				return
			}

			// validate host and domain of authorizer url
			if strings.TrimSuffix(stateObj.AuthorizerURL, "/") != envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyAuthorizerURL) {
				c.JSON(400, gin.H{"error": "invalid host url"})
				return
			}
		}

		// debug the request state
		if pusher := c.Writer.Pusher(); pusher != nil {
			// use pusher.Push() to do server push
			if err := pusher.Push("/app/build/bundle.js", nil); err != nil {
				log.Printf("Failed to push: %v", err)
			}
		}
		c.HTML(http.StatusOK, "app.tmpl", gin.H{
			"data": map[string]string{
				"authorizerURL":    stateObj.AuthorizerURL,
				"redirectURL":      stateObj.RedirectURL,
				"organizationName": envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyOrganizationName),
				"organizationLogo": envstore.EnvInMemoryStoreObj.GetStringStoreEnvVariable(constants.EnvKeyOrganizationLogo),
			},
		})
	}
}
