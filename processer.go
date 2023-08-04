package activitypub

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/Mushus/activitypub/internal"
	"github.com/Mushus/activitypub/lib/crypt"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type Processor struct {
	log          *zerolog.Logger
	urlResolver  *URLResolver
	remoteServer *RemoteServer
	accountStore AccountStore
	followStore  FollowStore
	host         string
}

func NewProcessor(
	config *Config,
	log *zerolog.Logger,
	urlResolver *URLResolver,
	remoteServer *RemoteServer,
	accountStore AccountStore,
	followStore FollowStore,
) *Processor {
	return &Processor{
		host:         urlResolver.host,
		log:          log,
		urlResolver:  urlResolver,
		remoteServer: remoteServer,
		accountStore: accountStore,
		followStore:  followStore,
	}
}

func (p *Processor) Webfinger(c context.Context, resource string) (*Account, error) {
	acct, err := parseAcctScheme(resource)
	if err != nil {
		return nil, fmt.Errorf("failed to parse acct: %w", err)
	}

	name := acct.preferredUsername
	host := acct.host

	if host != "" && host != p.host {
		return nil, ErrNotFound
	}

	return p.accountStore.FindByUsername(c, name)
}

type LocalAccountResult struct {
	Account *Account
	Actor   *Actor
}

func (p *Processor) GetMainKey(c context.Context, accountID string) (*internal.JSONMainKey, error) {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	publicKey, err := crypt.GeneratePuublicKeyPEM(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	actorURL := p.urlResolver.resolveActorURL(account.ID)
	publicKeyURL := p.urlResolver.resolveMainKeyURL(account.ID)

	return &internal.JSONMainKey{
		Context:           json.RawMessage(`["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"]`),
		ID:                actorURL,
		Type:              "Person",
		PreferredUsername: account.Username,
		PublicKey: internal.JSONPublicKey{
			ID:           publicKeyURL,
			Owner:        actorURL,
			PublicKeyPem: publicKey,
		},
	}, nil
}

func (p *Processor) GetLocalAccount(c context.Context, accountID string) (*internal.JSONActor, error) {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	publicKey, err := crypt.GeneratePuublicKeyPEM(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	actorURL := p.urlResolver.resolveActorURL(account.ID)
	publicKeyURL := p.urlResolver.resolveMainKeyURL(account.ID)
	inbox := p.urlResolver.resolveInboxURL(account.ID)
	outbox := p.urlResolver.resolveOutboxURL(account.ID)

	return &internal.JSONActor{
		Context:           json.RawMessage(`["https://www.w3.org/ns/activitystreams","https://w3id.org/security/v1"]`),
		ID:                actorURL,
		Type:              "Person",
		Discoverable:      true,
		Name:              account.Username,
		PreferredUsername: account.Username,
		URL:               actorURL,
		PublicKey: internal.JSONPublicKey{
			ID:           publicKeyURL,
			Owner:        actorURL,
			PublicKeyPem: publicKey,
		},
		Inbox:  inbox,
		Outbox: outbox,
	}, nil
}

type ViewResult struct {
	Actor      *Actor
	IsFollow   bool
	IsFollower bool
}

// View - アクターの状態を表示する
// acctStr は user@host の形式で指定する
// accountID を未指定として空文字が利用できる
func (p *Processor) View(c context.Context, accountID string, acctStr string) (*ViewResult, error) {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find account: %w", err)
	}

	acct, err := p.complementUserAddr(acctStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse acct: %w", err)
	}

	var isFollow bool
	var isFollower bool

	actor, err := p.findActor(c, account, acct)
	if err != nil {
		return nil, err
	}

	fromID := p.urlResolver.resolveActorURL(accountID)
	isFollow, err = p.followStore.IsFollowing(c, fromID, actor.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check following: %w", err)
	}

	isFollower, err = p.followStore.IsFollowing(c, actor.ID, fromID)
	if err != nil {
		return nil, fmt.Errorf("failed to check follower: %w", err)
	}

	return &ViewResult{
		Actor:      actor,
		IsFollow:   isFollow,
		IsFollower: isFollower,
	}, nil
}

// Login - ログインを行う
// 成功した場合アカウントのIDを返す
func (p *Processor) Login(c context.Context, email string, password string) (string, error) {
	account, err := p.accountStore.FindByEmail(c, email)
	if err != nil {
		return "", fmt.Errorf("failed to find account: %w", err)
	}

	if bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password)) != nil {
		return "", fmt.Errorf("invalid password")
	}

	return account.ID, nil
}

// Signup - サインアップを行う
// 成功した場合アカウントのIDを返す
func (p *Processor) Signup(c context.Context, email string, username string, password string) (string, error) {
	id := generateID()

	privateKey, err := crypt.GeneratePrivateKeyPEM()
	if err != nil {
		return "", fmt.Errorf("failed to generate private key: %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	account := &Account{
		ID:         id,
		Email:      email,
		Username:   username,
		Password:   string(hashedPassword),
		PrivateKey: privateKey,
	}

	err = p.accountStore.Save(c, account)
	if err != nil {
		return "", fmt.Errorf("failed to save account: %w", err)
	}

	return account.ID, nil
}

func (p *Processor) ReceiveInbox(c context.Context, accountID string, postReader io.Reader) error {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	post, err := io.ReadAll(postReader)
	if err != nil {
		return fmt.Errorf("failed to read post: %w", err)
	}

	var activity internal.JSONActivityType
	err = json.Unmarshal(post, &activity)
	if err != nil {
		return fmt.Errorf("failed to unmarshal activity: %w", err)
	}

	switch activity.Type {
	case "Follow":
		return p.receiveFollow(c, account, post)
	default:
		return fmt.Errorf("unsupported activity type: %s", activity.Type)
	}
}

func (p *Processor) ReceiveOutbox(c context.Context, accountID string, postReader io.Reader) error {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	post, err := io.ReadAll(postReader)
	if err != nil {
		return fmt.Errorf("failed to read post: %w", err)
	}

	var activity internal.JSONActivityType
	err = json.Unmarshal(post, &activity)
	if err != nil {
		return fmt.Errorf("failed to unmarshal activity: %w", err)
	}

	switch activity.Type {
	case "Follow":
		return p.receiveFollow(c, account, post)
	case "Undo":
		return p.receiveUndo(c, account, post)
	default:
		return fmt.Errorf("unsupported activity type: %s", activity.Type)
	}
}

func (p *Processor) receiveFollow(c context.Context, account *Account, post []byte) error {
	var follow internal.JSONFollow
	if err := json.Unmarshal(post, &follow); err != nil {
		return fmt.Errorf("failed to unmarshal follow: %w", err)
	}

	// TODO: main-keyのownerとactorが一致しているかどうかのチェック
	followerID := follow.Actor

	follower, err := p.remoteServer.GetActor(c, account, followerID)
	if err != nil {
		return fmt.Errorf("failed to get target: %w", err)
	}

	followsID := p.urlResolver.resolveActorURL(account.ID)
	if err := p.followStore.Follow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}

	// TODO: manual accept
	actorID := p.urlResolver.resolveActorURL(account.ID)
	accept := internal.JSONLDAccept{
		Context: []byte(`"https://www.w3.org/ns/activitystreams"`),
		JSONAccept: internal.JSONAccept{
			Type:   "Accept",
			ID:     p.urlResolver.resolveActorURL(account.ID) + "/a/" + generateID(), // TODO: 該当URL
			Actor:  actorID,
			Object: post,
		},
	}

	return p.remoteServer.PostInbox(c, account, follower.Inbox, accept)
}

func (p *Processor) receiveUndo(c context.Context, account *Account, post []byte) error {
	var undo internal.JSONUndo
	if err := json.Unmarshal(post, &undo); err != nil {
		return fmt.Errorf("failed to unmarshal undo: %w", err)
	}

	var activity internal.JSONActivityType
	if err := json.Unmarshal(undo.Object, &activity); err != nil {
		return fmt.Errorf("failed to unmarshal activity: %w", err)
	}

	switch activity.Type {
	case "Follow":
		return p.receiveUndoFollow(c, account, undo)
	default:
		return fmt.Errorf("unsupported activity type: %s", activity.Type)
	}
}

func (p *Processor) receiveUndoFollow(c context.Context, account *Account, undo internal.JSONUndo) error {
	var follow internal.JSONFollow
	if err := json.Unmarshal(undo.Object, &follow); err != nil {
		return fmt.Errorf("failed to unmarshal follow: %w", err)
	}

	followerID := follow.Actor
	followsID := p.urlResolver.resolveActorURL(account.ID)

	if err := p.followStore.Unfollow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}

// Follow - フォローを行う
// accountID はフォローするアカウントのID
// acctStr は user@host の形式で指定する
func (p *Processor) Follow(c context.Context, accountID string, acctStr string) error {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}

	acct, err := p.complementUserAddr(acctStr)
	if err != nil {
		return fmt.Errorf("failed to parse acct: %w", err)
	}

	actor, err := p.findActor(c, account, acct)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	// リモートユーザーには通知が必要
	if actor.Host == p.host {
		return p.followLocal(c, account, acct)
	}
	return p.followRemote(c, account, acct)
}

func (p *Processor) followRemote(c context.Context, account *Account, acct *userAddr) error {
	actor, err := p.findRemoteActor(c, account, acct)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	followerID := p.urlResolver.resolveActorURL(account.ID)
	followsID := actor.ID

	activityID := GenerateSortableID()
	followActivityID := p.urlResolver.resolveActivityURL(activityID)

	toJSON, err := json.Marshal(followsID)
	if err != nil {
		return fmt.Errorf("failed to marshal followID: %w", err)
	}

	// TODO: アクティビティの保存

	followBody := internal.JSONLDFollow{
		Context: []byte(`"https://www.w3.org/ns/activitystreams"`),
		JSONFollow: internal.JSONFollow{
			Type:   "Follow",
			ID:     followActivityID,
			Actor:  followerID,
			Object: toJSON,
			To:     toJSON,
		},
	}

	if err := p.remoteServer.PostInbox(c, account, actor.Inbox, followBody); err != nil {
		return fmt.Errorf("failed to post inbox: %w", err)
	}

	if err := p.followStore.Follow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}

	return nil
}

func (p *Processor) followLocal(c context.Context, account *Account, acct *userAddr) error {
	actor, err := p.findLocalActor(c, acct)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	followerID := p.urlResolver.resolveActorURL(account.ID)
	followsID := actor.ID

	if err := p.followStore.Follow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to follow: %w", err)
	}

	return nil
}

// Unfollow - フォローを解除する
// accountID はフォローするアカウントのID
// acctStr は user@host の形式で指定する
func (p *Processor) Unfollow(c context.Context, accountID string, acctStr string) error {
	account, err := p.accountStore.Find(c, accountID)
	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}

	acct, err := p.complementUserAddr(acctStr)
	if err != nil {
		return fmt.Errorf("failed to parse acct: %w", err)
	}

	if acct.host == p.host {
		return p.unfollowLocal(c, account, acct)
	} else {
		return p.unfollowRemote(c, account, acct)
	}
}

// unfollowLocal - フォローを解除する
// accountID はフォローするアカウントのID
// acctStr は user@host の形式で指定する
func (p *Processor) unfollowLocal(c context.Context, account *Account, acct *userAddr) error {
	actor, err := p.findLocalActor(c, acct)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	followerID := p.urlResolver.resolveActorURL(account.ID)
	followsID := actor.ID

	if err := p.followStore.Unfollow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}

// unfollowRemote - フォローを解除する
// accountID はフォローするアカウントのID
// acctStr は user@host の形式で指定する
func (p *Processor) unfollowRemote(c context.Context, account *Account, acct *userAddr) error {
	actor, err := p.findRemoteActor(c, account, acct)
	if err != nil {
		return fmt.Errorf("failed to find actor: %w", err)
	}

	followerID := p.urlResolver.resolveActorURL(account.ID)
	followsID := actor.ID

	// TODO: Followの取り消しを送信する

	if err := p.followStore.Unfollow(c, followerID, followsID); err != nil {
		return fmt.Errorf("failed to unfollow: %w", err)
	}

	return nil
}

// findUser - ユーザーを検索する
// acctStr は user@host の形式で指定する
// account は nil でも良い
// 正常終了でユーザーが見つからなかったときは ErrNotFound を返す
func (p *Processor) findActor(c context.Context, account *Account, acct *userAddr) (*Actor, error) {
	if acct.host == p.host {
		return p.findLocalActor(c, acct)
	} else {
		jsonActor, err := p.findRemoteActor(c, account, acct)
		if err != nil {
			return nil, fmt.Errorf("failed to find remote actor: %w", err)
		}
		return createActorFromJSON(jsonActor, acct.host), nil
	}
}

// findLocalActor - ユーザーを検索する
// acctStr は user@host の形式で指定する
// 正常終了でユーザーが見つからなかったときは ErrNotFound を返す
func (p *Processor) findLocalActor(c context.Context, acct *userAddr) (*Actor, error) {
	actor, err := p.accountStore.FindByUsername(c, acct.preferredUsername)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			p.log.Debug().Msg("user not found")
		} else {
			return nil, fmt.Errorf("failed to find actor: %w", err)
		}
	}
	return p.createActorFromAccount(actor)
}

// findLocalActor - ユーザーを検索する
// acctStr は user@host の形式で指定する
// account は nil でも良い
// 正常終了でユーザーが見つからなかったときは ErrNotFound を返す
func (p *Processor) findRemoteActor(c context.Context, account *Account, acct *userAddr) (*internal.JSONActor, error) {
	if account == nil {
		return nil, ErrNotFound
	}

	resource := fmt.Sprintf("acct:%s@%s", acct.preferredUsername, acct.host)
	webfinger, err := p.remoteServer.GetWebfinger(c, acct.host, resource)
	if err != nil {
		return nil, fmt.Errorf("failed to get webfinger: %w", err)
	}

	actorID := findActorIDFromWebfinger(webfinger)
	if actorID == "" {
		return nil, fmt.Errorf("failed to find actorID from webfinger")
	}

	return p.remoteServer.GetActor(c, account, actorID)
}

// complementUserAddr - 完全なユーザー名を解析する
func (p *Processor) complementUserAddr(acctStr string) (*userAddr, error) {
	acct, err := parseUserAddr(acctStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse acct: %w", err)
	}

	if acct.host == "" {
		acct.host = p.host
	}

	return acct, nil
}

func (p *Processor) createActorFromAccount(account *Account) (*Actor, error) {
	publicKey, err := crypt.GeneratePuublicKeyPEM(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate public key: %w", err)
	}

	actorID := p.urlResolver.resolveActorURL(account.ID)
	return &Actor{
		ID:        actorID,
		Username:  account.Username,
		Host:      p.host,
		PublicKey: publicKey,
	}, nil
}

func createActorFromJSON(actor *internal.JSONActor, host string) *Actor {
	return &Actor{
		ID:        actor.ID,
		Username:  actor.PreferredUsername,
		Host:      host,
		PublicKey: actor.PublicKey.PublicKeyPem,
	}
}
