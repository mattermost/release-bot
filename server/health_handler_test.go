package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mattermost/release-bot/version"
	"github.com/stretchr/testify/assert"
)

func TestHealthzRoute(t *testing.T) {
	handler := newHeathHandler()
	req := httptest.NewRequest(http.MethodGet, healthHandlerDefaultRoute, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	defer res.Body.Close()
	assert.Equal(t, "application/json;charset=utf-8", res.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	version := &version.Info{}
	err = json.Unmarshal(data, version)
	assert.Nil(t, err)
	assert.Equal(t, "Release-Bot", version.Name)
	assert.Equal(t, "dev", version.Hash)
	assert.Equal(t, "dev", version.Version)
}
