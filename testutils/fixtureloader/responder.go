package fixtureloader

import (
	"encoding/json"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func (l *Loader) CreateJSONResponder(t *testing.T, status int, path string) httpmock.Responder {
	resp := httpmock.NewStringResponse(status, l.LoadString(t, path))
	resp.Header.Set("Content-Type", "application/json")
	return httpmock.ResponderFromResponse(resp)
}

func (l *Loader) CreateJSONResponderWithPayload(t *testing.T, status int, payload interface{}) httpmock.Responder {
	b, err := json.Marshal(payload)
	require.NoError(t, err)

	resp := httpmock.NewStringResponse(status, string(b))
	resp.Header.Set("Content-Type", "application/json")
	return httpmock.ResponderFromResponse(resp)
}
