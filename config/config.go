package config

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Server    HTTPConfig       `mapstructure:"server"`
	Queue     QueueConfig      `mapstructure:"queue"`
	Github    GithubConfig     `mapstructure:"github"`
	Pipelines []PipelineConfig `mapstructure:"pipelines"`
}

type HTTPConfig struct {
	BaseURL string `mapstructure:"base_url"`
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

type QueueConfig struct {
	Limit   int `mapstructure:"limit"`
	Workers int `mapstructure:"workers"`
}
type GithubConfig struct {
	IntegrationID int64  `mapstructure:"integration_id"`
	WebhookSecret string `mapstructure:"webhook_secret"`
	PrivateKey    string `mapstructure:"private_key"`
}

type PipelineConfig struct {
	Organization string              `mapstructure:"organization"`
	Repository   string              `mapstructure:"repository"`
	Workflow     string              `mapstructure:"workflow"`
	TargetBranch string              `mapstructure:"targetBranch"`
	Timeout      string              `mapstructure:"timeout"`
	SleepSeconds int64               `mapstructure:"sleepSeconds"`
	Context      string              `mapstructure:"context"`
	Conditions   []PipelineCondition `mapstructure:"conditions"`
}

type PipelineCondition struct {
	Repository string   `mapstructure:"repository"`
	Webhook    []string `mapstructure:"webhook"`
	Workflow   string   `mapstructure:"workflow"`
	Type       string   `mapstructure:"type"`
	Fork       bool     `mapstructure:"fork"`
	Status     string   `mapstructure:"status"`
	Conclusion string   `mapstructure:"conclusion"`
	Name       string   `mapstructure:"name"`
}

func ReadConfig(filename string, paths ...string) (*Config, error) {
	if len(paths) == 0 {
		return nil, errors.New("Please provide configuration file location.")
	}
	log.Info("Loading configuration.")

	viper.SetConfigName(filename)
	viper.SetConfigType("yaml")
	for i := range paths {
		viper.AddConfigPath(paths[i])
	}

	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		return nil, errors.Wrapf(err, "failed reading server config file: config/config.yaml")
	}

	var c Config

	if err := viper.Unmarshal(&c); err != nil {
		return nil, errors.Wrap(err, "failed parsing configuration file")
	}
	return &c, nil
}
