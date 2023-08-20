package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"emperror.dev/errors"
	"github.com/Mushus/activitypub"
	"github.com/Mushus/activitypub/lib/array"
	"github.com/Mushus/activitypub/sqlite/ent"
	"github.com/Mushus/activitypub/sqlite/ent/account"
	"github.com/Mushus/activitypub/sqlite/ent/follow"
	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	cli *ent.Client
}

func NewSQLite() (*SQLite, error) {
	cli, err := ent.Open("sqlite3", "./database.db?_fk=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open ent client: %w", errors.WithStack(err))
	}

	ctx := context.Background()
	if err := cli.Schema.Create(ctx); err != nil {
		return nil, fmt.Errorf("failed creating schema resources: %w", errors.WithStack(err))
	}

	return &SQLite{cli: cli}, nil
}

// account

type AccountDB struct {
	*SQLite
}

func NewAccountDB(db *SQLite) activitypub.AccountStore {
	return &AccountDB{SQLite: db}
}

func (d *AccountDB) Save(c context.Context, acc *activitypub.Account) error {
	_, err := d.cli.Account.Create().
		SetID(acc.ID).
		SetUsername(acc.Username).
		SetEmail(acc.Email).
		SetPassword(acc.Password).
		SetPrivateKey(acc.PrivateKey).
		Save(c)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", errors.WithStack(err))
	}
	return nil
}

func (d *AccountDB) Find(c context.Context, id string) (*activitypub.Account, error) {
	account, err := d.cli.Account.Get(c, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, activitypub.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", errors.WithStack(err))
	}
	return toAccount(account)
}

func (d *AccountDB) FindByEmail(c context.Context, email string) (*activitypub.Account, error) {
	account, err := d.cli.Account.Query().
		Where(account.Email(email)).
		First(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, activitypub.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", errors.WithStack(err))
	}
	return toAccount(account)
}

func (d *AccountDB) FindByUsername(c context.Context, username string) (*activitypub.Account, error) {
	account, err := d.cli.Account.Query().
		Where(account.Username(username)).
		First(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, activitypub.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", errors.WithStack(err))
	}
	return toAccount(account)
}

func toAccount(account *ent.Account) (*activitypub.Account, error) {
	return &activitypub.Account{
		ID:         account.ID,
		Username:   account.Username,
		Email:      account.Email,
		Password:   account.Password,
		PrivateKey: account.PrivateKey,
	}, nil
}

// follow

type FollowDB struct {
	*SQLite
}

func NewFollowDB(db *SQLite) activitypub.FollowStore {
	return &FollowDB{SQLite: db}
}

func (d *FollowDB) RequestFollow(c context.Context, fromID string, toID string) error {
	err := d.cli.Follow.Create().
		SetID(activitypub.GenerateSortableID()).
		SetFromID(fromID).
		SetToID(toID).
		SetStatus(followStatusPending.toValue()).
		OnConflict().
		UpdateNewValues().
		Exec(c)
	if err != nil {
		return fmt.Errorf("failed to create follow: %w", errors.WithStack(err))
	}
	return nil
}

func (d *FollowDB) Follow(c context.Context, fromID string, toID string) error {
	err := d.cli.Follow.Create().
		SetID(activitypub.GenerateSortableID()).
		SetFromID(fromID).
		SetToID(toID).
		SetStatus(followStatusFollowing.toValue()).
		OnConflict().
		UpdateNewValues().
		Exec(c)
	if err != nil {
		return fmt.Errorf("failed to create follow: %w", errors.WithStack(err))
	}
	return nil
}

func (d *FollowDB) Unfollow(c context.Context, fromID string, toID string) error {
	_, err := d.cli.Follow.Delete().
		Where(follow.FromID(fromID), follow.ToID(toID)).
		Exec(c)
	if err != nil {
		return fmt.Errorf("failed to delete follow: %w", errors.WithStack(err))
	}
	return nil
}

func (d *FollowDB) FindFollowStatus(c context.Context, fromID string, toID string) (activitypub.FollowStatus, error) {
	status, err := d.cli.Follow.Query().
		Where(
			follow.FromID(fromID),
			follow.ToID(toID),
		).
		First(c)
	if err != nil {
		if ent.IsNotFound(err) {
			return activitypub.FollowStatusUnfollowing, nil
		}
		return activitypub.FollowStatusUnknown, fmt.Errorf("failed to get follow: %w", errors.WithStack(err))
	}
	return activitypub.FindFollowStatus(status.Status), nil
}

func (d *FollowDB) ListFollowers(c context.Context, id string) ([]string, error) {
	followers, err := d.cli.Follow.Query().
		Where(
			follow.ToID(id),
			follow.Status(followStatusFollowing.toValue()),
		).
		All(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get follows: %w", errors.WithStack(err))
	}
	return array.Map(followers, func(follow *ent.Follow) string {
		return follow.FromID
	}), nil
}

func (d *FollowDB) ListFollows(c context.Context, id string) ([]string, error) {
	follows, err := d.cli.Follow.Query().
		Where(
			follow.FromID(id),
			follow.Status(followStatusFollowing.toValue()),
		).
		All(c)
	if err != nil {
		return nil, fmt.Errorf("failed to get follows: %w", errors.WithStack(err))
	}
	return array.Map(follows, func(follow *ent.Follow) string {
		return follow.ToID
	}), nil
}

// avtivity

// type ActivityDB struct {
// 	*SQLite
// }

// func NewActivityDB(db *SQLite) activitypub.ActivityStore {
// 	return &ActivityDB{SQLite: db}
// }

// func (d *ActivityDB) Save(c context.Context, activity *activitypub.Activity) error {
// 	d.cli.Activity.Create().
// 		SetID(activity.ID).
// 		SetActivityID(activity.ActivityID).
// 		SetJSON(activity.JSON).
// 		Exec(c)
// 	return nil
// }

// func (d *ActivityDB) Find(c context.Context, id string) (*activitypub.Activity, error) {
// 	act, err := d.cli.Activity.Query().
// 		Where(activity.ID(id)).
// 		First(c)
// 	if err != nil {
// 		if ent.IsNotFound(err) {
// 			return nil, activitypub.ErrNotFound
// 		}
// 		return nil, fmt.Errorf("failed to get activity: %w", errors.WithStack(err))
// 	}
// 	return toActivity(act)
// }

// func toActivity(act *ent.Activity) (*activitypub.Activity, error) {
// 	return &activitypub.Activity{
// 		ID:         act.ID,
// 		ActivityID: act.ActivityID,
// 		JSON:       act.JSON,
// 	}, nil
// }

// session

type Sqlite3Session struct {
	sess *scs.SessionManager
	db   *sql.DB
}

func NewSession() (activitypub.Session, error) {
	db, err := sql.Open("sqlite3", "session.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			token TEXT PRIMARY KEY,
			data BLOB NOT NULL,
			expiry REAL NOT NULL
		);

		CREATE INDEX IF NOT EXISTS sessions_expiry_idx ON sessions(expiry);`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session table: %w", errors.WithStack(err))
	}

	sess := scs.New()
	sess.Store = sqlite3store.New(db)
	sess.Lifetime = 30 * 24 * time.Hour
	sess.Cookie.Name = "session_id"
	sess.Cookie.HttpOnly = true
	sess.Cookie.Persist = true
	sess.Cookie.SameSite = http.SameSiteStrictMode
	sess.Cookie.Secure = true

	return &Sqlite3Session{
		sess: sess,
	}, nil
}

func (s *Sqlite3Session) Close() error {
	return s.db.Close()
}

func (s *Sqlite3Session) Set(c context.Context, key string, value any) {
	s.sess.Put(c, key, value)
}

func (s *Sqlite3Session) Get(c context.Context, key string) any {
	return s.sess.Get(c, key)
}

func (s *Sqlite3Session) Delete(c context.Context, key string) {
	s.sess.Remove(c, key)
}

func (s *Sqlite3Session) Clear(c context.Context) {
	s.sess.Clear(c)
}

func (s *Sqlite3Session) Middleware(next http.Handler) http.Handler {
	return s.sess.LoadAndSave(next)
}
