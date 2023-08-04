package activitypub

import "context"

type AccountStore interface {
	Find(c context.Context, id string) (*Account, error)
	FindByEmail(c context.Context, email string) (*Account, error)
	FindByUsername(c context.Context, username string) (*Account, error)
	Save(c context.Context, account *Account) error
}

type FollowStore interface {
	Follow(c context.Context, fromID string, toID string) error
	Unfollow(c context.Context, fromID string, toID string) error
	IsFollowing(c context.Context, fromID string, toID string) (bool, error)
	ListFollowers(c context.Context, id string) ([]string, error)
	ListFollows(c context.Context, id string) ([]string, error)
}
