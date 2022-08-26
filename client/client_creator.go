package client

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v45/github"
	lru "github.com/hashicorp/golang-lru"
	"github.com/mattermost/release-bot/version"
	"github.com/pkg/errors"
)

const (
	ClientCacheSize = 64
)

type ClientCache interface {
	Get(installationID int64) (*github.Client, error)
}

type clientCache struct {
	appID          int64
	privateKeyFile string
	userAgent      string
	cache          *lru.Cache
	transport      http.RoundTripper
}

func NewClientCache(appID int64, privateKeyFile string) (ClientCache, error) {
	cache, err := lru.New(ClientCacheSize)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cache")
	}
	version := version.Full()
	return &clientCache{
		appID:          appID,
		privateKeyFile: privateKeyFile,
		cache:          cache,
		transport:      http.DefaultTransport,
		userAgent:      fmt.Sprintf("%s/%s", version.Name, version.Version),
	}, nil
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
