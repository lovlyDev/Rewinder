package plugins

import (
	"encoding/json"

	"Rewinder/internal/state"
)

func jsonRoundTrip(in state.AppState) (state.AppState, error) {
	var out state.AppState
	b, err := json.Marshal(in)
	if err != nil {
		return out, err
	}
	err = json.Unmarshal(b, &out)
	return out, err
}
