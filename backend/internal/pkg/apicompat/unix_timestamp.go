package apicompat

import (
	"encoding/json"
	"math"
)

// unixSeconds decodes a unix timestamp that some OpenAI-compatible upstreams
// send as a float (1795533445.0) instead of an integer.
type unixSeconds int64

func (u *unixSeconds) UnmarshalJSON(data []byte) error {
	var n json.Number
	if err := json.Unmarshal(data, &n); err != nil || n == "" {
		return err
	}
	if i, err := n.Int64(); err == nil {
		*u = unixSeconds(i)
		return nil
	}
	f, err := n.Float64()
	if err != nil {
		return err
	}
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	*u = unixSeconds(int64(f))
	return nil
}

func (r *ResponsesResponse) UnmarshalJSON(data []byte) error {
	type alias ResponsesResponse
	aux := struct {
		*alias
		CreatedAt unixSeconds `json:"created_at"`
	}{alias: (*alias)(r)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.CreatedAt = int64(aux.CreatedAt)
	return nil
}

func (r *ChatCompletionsResponse) UnmarshalJSON(data []byte) error {
	type alias ChatCompletionsResponse
	aux := struct {
		*alias
		Created unixSeconds `json:"created"`
	}{alias: (*alias)(r)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.Created = int64(aux.Created)
	return nil
}

func (c *ChatCompletionsChunk) UnmarshalJSON(data []byte) error {
	type alias ChatCompletionsChunk
	aux := struct {
		*alias
		Created unixSeconds `json:"created"`
	}{alias: (*alias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	c.Created = int64(aux.Created)
	return nil
}
