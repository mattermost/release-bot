package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/mattermost/release-bot/client"
	"github.com/mattermost/release-bot/store"
	"github.com/stretchr/testify/assert"
)

func TestCheckRunUpdateRouteInvalidMethod(t *testing.T) {
	eventContextStore := store.NewEventContextStore()

	cc, _ := client.NewClientCache(int64(100), "Private Key File")
	handler := newCheckRunUpdateHandler(cc, eventContextStore)
	t.Run("Method Not Allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, updateCheckRunDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "405 Method Not Allowed", res.Status)
	})
	t.Run("Bad Request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, updateCheckRunDefaultRoute, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		res := w.Result()
		assert.Equal(t, "400 Bad Request", res.Status)
	})
}

func TestCheckRunUpdateRoute(t *testing.T) {
	token := "token"
	eventContextStore := store.NewEventContextStore()
	eventContext := createWorkflowRunEvent(t)
	eventContextStore.Store(eventContext, token)

	handler := newCheckRunUpdateHandler(&mockClientCache{}, eventContextStore)
	request, _ := os.Open("testdata/check_run_update_request.json")

	req := httptest.NewRequest(http.MethodPost, updateCheckRunDefaultRoute, request)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	res := w.Result()
	assert.Equal(t, "200 OK", res.Status)
	defer res.Body.Close()
	assert.Equal(t, "application/json;charset=utf-8", res.Header.Get("Content-Type"))
	data, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	ccr := &CreateCheckRunResponse{}
	err = json.Unmarshal(data, ccr)
	assert.Nil(t, err)
	assert.Equal(t, int64(100), *ccr.ID)
}
