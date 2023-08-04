package activitypub

import (
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid/v2"
)

// generateID - IDの生成
func generateID() string {
	id := uuid.New()
	idStr := strings.ReplaceAll(id.String(), "-", "")
	return idStr
}

func GenerateSortableID() string {
	t := time.Now()
	entropy := ulid.Monotonic(rand.New(rand.NewSource(t.UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	return id.String()
}

type Account struct {
	ID         string
	Username   string
	Email      string
	Password   string
	PrivateKey string
}

type Actor struct {
	ID        string
	Username  string
	Host      string
	PublicKey string
}
