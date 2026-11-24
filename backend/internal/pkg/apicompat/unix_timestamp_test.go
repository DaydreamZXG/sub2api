package apicompat

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFloatTimestampsDecode(t *testing.T) {
	var ev ResponsesStreamEvent
	require.NoError(t, json.Unmarshal([]byte(`{"type":"response.completed","response":{"id":"r","created_at":1795533445.0,"model":"m","status":"completed","output":[]}}`), &ev))
	require.Equal(t, int64(1795533445), ev.Response.CreatedAt)
	require.Equal(t, "m", ev.Response.Model)

	var resp ResponsesResponse
	require.NoError(t, json.Unmarshal([]byte(`{"id":"r","created_at":1795533445}`), &resp))
	require.Equal(t, int64(1795533445), resp.CreatedAt)
	require.NoError(t, json.Unmarshal([]byte(`{"id":"r2","created_at":null}`), &resp))
	require.Equal(t, "r2", resp.ID)

	var chat ChatCompletionsResponse
	require.NoError(t, json.Unmarshal([]byte(`{"id":"c","created":1795533445.5,"model":"m","choices":[]}`), &chat))
	require.Equal(t, int64(1795533445), chat.Created)

	var chunk ChatCompletionsChunk
	require.NoError(t, json.Unmarshal([]byte(`{"id":"c","created":1795533445.0,"choices":[]}`), &chunk))
	require.Equal(t, int64(1795533445), chunk.Created)

	require.Error(t, json.Unmarshal([]byte(`{"created_at":"x"}`), &resp))
}
