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
	RequestFollow(c context.Context, fromID string, toID string) error
	Unfollow(c context.Context, fromID string, toID string) error
	FindFollowStatus(c context.Context, fromID string, toID string) (FollowStatus, error)
	ListFollowers(c context.Context, id string) ([]string, error)
	ListFollows(c context.Context, id string) ([]string, error)
}

// type ActivityStore interface {
// 	Save(c context.Context, activity *Activity) error
// 	Find(c context.Context, id string) (*Activity, error)
// }
