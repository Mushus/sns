package sqlite

// followStatus is a status of follow.
type followStatus int

const (
	// followStatusUnknown is unknown status.
	followStatusUnknown followStatus = iota
	// followStatusFollowing is following status.
	followStatusFollowing
	// FollowStatusFollowed is follow request status.
	followStatusPending
)

func (s followStatus) toValue() int {
	return int(s)
}
