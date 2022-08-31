package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v45/github"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mattermost/release-bot/config"
	"github.com/mattermost/release-bot/version"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	ClientCacheSize = 128
)

type GithubClientManager interface {
	Get(installationID int64) (*github.Client, error)
	CreateToken(repository string, runID int64, installationID int64) (AccessToken, error)
	RevokeToken(repository string, runID int64) error
}

type AccessToken interface {
	GetInstallationID() int64
	IsExpired() bool
	GetToken() string
}

type clientCache struct {
	appID              int64
	privateKeyFile     string
	userAgent          string
	cache              *lru.Cache
	transport          http.RoundTripper
	appClient          *github.Client
	installationTokens map[string]AccessToken
}

type accessToken struct {
	installationID int64
	token          string
	expiresAt      time.Time
}

func BuildFromConfig(config *config.Config) (GithubClientManager, error) {
	cache, err := lru.New(ClientCacheSize)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cache")
	}
	version := version.Full()
	itr, err := ghinstallation.NewAppsTransportKeyFromFile(http.DefaultTransport, config.Github.IntegrationID, config.Github.PrivateKey)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not initialize GitHub client! Check configuration values.")
	}
	return New(
		config.Github.IntegrationID,
		config.Github.PrivateKey,
		cache,
		http.DefaultTransport,
		fmt.Sprintf("%s/%s", version.Name, version.Version),
		github.NewClient(&http.Client{Transport: itr}),
	), nil
}

func New(
	appID int64,
	privateKeyFile string,
	cache *lru.Cache,
	transport http.RoundTripper,
	userAgent string,
	appClient *github.Client,
) GithubClientManager {
	return &clientCache{
		appID:              appID,
		privateKeyFile:     privateKeyFile,
		cache:              cache,
		transport:          transport,
		userAgent:          userAgent,
		appClient:          appClient,
		installationTokens: make(map[string]AccessToken),
	}
}

func (cc *clientCache) Get(installationID int64) (*github.Client, error) {
	cli, ok := cc.cache.Get(installationID)
	if ok {
		return cli.(*github.Client), nil
	}
	itr, err := ghinstallation.NewKeyFromFile(
		cc.transport,
		cc.appID,
		installationID,
		cc.privateKeyFile,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "Can not initialize GitHub client! Check configuration values.")
	}
	client := github.NewClient(&http.Client{Transport: itr})
	client.UserAgent = cc.userAgent
	cc.cache.Add(installationID, client)
	return client, nil
}

func (cc *clientCache) CreateToken(repository string, runID int64, installationID int64) (AccessToken, error) {
	log.WithFields(log.Fields{
		"repository":      repository,
		"run_id":          runID,
		"installation_id": installationID,
	}).Info("Github Installation Token requested")
	mapKey := fmt.Sprintf("%s-%v", repository, runID)
	token, found := cc.installationTokens[mapKey]
	if found && !token.IsExpired() && token.GetInstallationID() == installationID {
		log.Info("Using non-expired repository github token")
		return token, nil
	}
	ghToken, _, err := cc.appClient.Apps.CreateInstallationToken(context.Background(), installationID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Can not create access token!")
	}
	token = newAccessToken(installationID, ghToken.GetToken(), ghToken.GetExpiresAt())
	cc.installationTokens[mapKey] = token
	return token, nil
}

func (cc *clientCache) RevokeToken(repository string, runID int64) error {
	log.
		WithFields(log.Fields{
			"repository": repository,
			"run_id":     runID,
		}).
		Info("Will revoke token for workflow run.")
	mapKey := fmt.Sprintf("%s-%v", repository, runID)
	token, found := cc.installationTokens[mapKey]
	if !found {
		log.WithFields(log.Fields{
			"repository": repository,
			"run_id":     runID,
		}).Info("No token is created.")
		return nil
	}
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", "https://api.github.com/installation/token", nil)
	if err != nil {
		log.
			WithFields(log.Fields{
				"repository": repository,
				"run_id":     runID,
			}).
			WithError(err).
			Error("Can not create delete token request")
		return errors.Wrap(err, "Can not create delete token request!")
	}
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token.GetToken()))

	resp, err := client.Do(req)
	if err != nil {
		log.
			WithFields(log.Fields{
				"repository": repository,
				"run_id":     runID,
			}).
			WithError(err).
			Error("Can not invoke delete token request!")
		return errors.Wrap(err, "Can not invoke delete token request!")
	}
	if resp.StatusCode != 204 {
		log.
			WithFields(log.Fields{
				"repository": repository,
				"run_id":     runID,
			}).
			WithError(err).
			Error("Github Token invalidation error")
		return errors.New("Token invalidation error!")
	}
	log.
		WithFields(log.Fields{
			"repository": repository,
			"run_id":     runID,
		}).
		Info("Token invalidated!")
	return nil
}

func newAccessToken(installationID int64, token string, expiresAt time.Time) AccessToken {
	return &accessToken{
		installationID: installationID,
		token:          token,
		expiresAt:      expiresAt,
	}
}

func (t *accessToken) IsExpired() bool {
	return t.expiresAt.Before(time.Now().Add(-1 * time.Minute))
}
func (t *accessToken) GetToken() string {
	return t.token
}
func (t *accessToken) GetInstallationID() int64 {
	return t.installationID
}
