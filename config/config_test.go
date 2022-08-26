package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguration(t *testing.T) {
	t.Run("No configuration location", func(t *testing.T) {
		config, err := ReadConfig("config_sample")
		assert.Nil(t, config)
		assert.Error(t, err)
		assert.Equal(t, "Please provide configuration file location.", err.Error())
	})
	t.Run("Missing configuration file", func(t *testing.T) {
		config, err := ReadConfig("", "test")
		assert.Nil(t, config)
		assert.Error(t, err)
	})
	t.Run("Test Configuration", func(t *testing.T) {
		config, err := ReadConfig("config_sample", "testdata")
		assert.NotNil(t, config)
		assert.Nil(t, err)
		assert.Equal(t, "https://api.github.com", config.Github.ApiURL)
		assert.Equal(t, int64(12345), config.Github.IntegrationID)
		assert.Equal(t, "certs/private_key.pem", config.Github.PrivateKey)
		assert.Equal(t, "N/A", config.Github.WebhookSecret)
		assert.Equal(t, 1, len(config.Pipelines))
		assert.Equal(t, "mattermost", config.Pipelines[0].Organization)
		assert.Equal(t, "******", config.Pipelines[0].Repository)
		assert.Equal(t, "docker.yaml", config.Pipelines[0].Workflow)
		assert.Equal(t, 10000, config.Queue.Limit)
		assert.Equal(t, 10, config.Queue.Workers)
		assert.Equal(t, "0.0.0.0", config.Server.Address)
		assert.Equal(t, "https://test.url.com", config.Server.BaseURL)
		assert.Equal(t, 8080, config.Server.Port)
	})
}
