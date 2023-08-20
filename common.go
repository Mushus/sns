package activitypub

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	SoftwareName string `envconfig:"SOFTWARE_NAME" default:"activitypub"`
	Host         string `envconfig:"HOST" default:"localhost:8080"`
	Port         int    `envconfig:"PORT" default:"8080"`
	Https        bool   `envconfig:"HTTPS" default:"false"`
}

func ParseConfig() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("activitypub", &cfg); err != nil {
		return nil, fmt.Errorf("cannot parse config: %w", err)
	}
	return &cfg, nil
}

// type Activity struct {
// 	ID         string `json:"id"`
// 	ActivityID string `json:"activity_id"`
// 	JSON       string `json:"json"`
// }

type FollowStatus int

const (
	FollowStatusUnknown     FollowStatus = -1
	FollowStatusFollowing   FollowStatus = 0
	FollowStatusPending     FollowStatus = 1
	FollowStatusUnfollowing FollowStatus = 2
)

func (s FollowStatus) Value() int {
	return int(s)
}

func FindFollowStatus(v int) FollowStatus {
	switch v {
	case FollowStatusFollowing.Value():
		return FollowStatusFollowing
	case FollowStatusPending.Value():
		return FollowStatusPending
	case FollowStatusUnfollowing.Value():
		return FollowStatusUnfollowing
	default:
		return FollowStatusUnknown
	}
}
