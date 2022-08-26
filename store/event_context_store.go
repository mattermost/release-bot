package store

import (
	"fmt"
	"time"

	"github.com/akyoto/cache"
	"github.com/mattermost/release-bot/model"
)

var cacheExpireInterval time.Duration
var itemExpireDuration time.Duration

func init() {
	cacheExpireInterval = 10 * time.Minute
	itemExpireDuration = 6 * time.Hour
}

type EventContextStore interface {
	Store(context model.EventContext, token string)
	Get(token string) (model.EventContext, error)
}

type eventContextStore struct {
	ItemDuration *time.Duration
	Cache        *cache.Cache
}

func NewEventContextStore() EventContextStore {
	return &eventContextStore{
		ItemDuration: &itemExpireDuration,
		Cache:        cache.New(cacheExpireInterval),
	}
}

func (store *eventContextStore) Store(eventContext model.EventContext, token string) {
	store.Cache.Set(token, eventContext, *store.ItemDuration)
}

func (store *eventContextStore) Get(token string) (model.EventContext, error) {
	context, found := store.Cache.Get(token)
	if !found {
		return nil, fmt.Errorf("not found")
	}
	return context.(model.EventContext), nil
}
