package contract

import "time"

const (
	// DPConfigRefreshChannel is the Redis Pub/Sub channel used by CP to notify DP
	// about snapshot/meta refresh events.
	DPConfigRefreshChannel = "dp:config:refresh"
)

type DPConfigRefreshEvent struct {
	Resource  string `json:"resource"`
	Version   string `json:"version,omitempty"`
	ActiveKey string `json:"active_key,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

func NewDPConfigRefreshEvent(resource, version, activeKey string) DPConfigRefreshEvent {
	return DPConfigRefreshEvent{
		Resource:  resource,
		Version:   version,
		ActiveKey: activeKey,
		Timestamp: time.Now().Unix(),
	}
}
