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
	CreateToken(installationID int64) (AccessToken, error)
}

type AccessToken interface {
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
	installationTokens map[int64]AccessToken
}

type accessToken struct {
	token     string
	expiresAt time.Time
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
		installationTokens: make(map[int64]AccessToken),
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

func (cc *clientCache) CreateToken(installationID int64) (AccessToken, error) {
	token, found := cc.installationTokens[installationID]
	if found && !token.IsExpired() {
		log.Info("Using non-expired repository github token")
		return token, nil
	}
	ghToken, _, err := cc.appClient.Apps.CreateInstallationToken(context.Background(), installationID, &github.InstallationTokenOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Can not create access token!")
	}
	token = newAccessToken(ghToken.GetToken(), ghToken.GetExpiresAt())
	cc.installationTokens[installationID] = token
	return token, nil
}

func newAccessToken(token string, expiresAt time.Time) AccessToken {
	return &accessToken{
		token:     token,
		expiresAt: expiresAt,
	}
}

func (t *accessToken) IsExpired() bool {
	return t.expiresAt.Before(time.Now().Add(-1 * time.Minute))
}
func (t *accessToken) GetToken() string {
	return t.token
}
