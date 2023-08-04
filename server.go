package activitypub

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	"github.com/Mushus/activitypub/internal"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ironstar-io/chizerolog"
	"github.com/rs/zerolog"
)

const InternalServerError = "internal server error"
const sessionKey = "session_id"

// interface

type Session interface {
	Close() error
	Set(c context.Context, key string, value any)
	Get(c context.Context, key string) any
	Delete(c context.Context, key string)
	Clear(c context.Context)
	Middleware(next http.Handler) http.Handler
}

// Server

type Server struct {
	handler *Handler
	port    int
}

func NewServer(cfg *Config, handler *Handler) (*Server, error) {
	return &Server{
		handler: handler,
		port:    cfg.Port,
	}, nil
}

func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	return http.ListenAndServe(addr, s.handler)
}

// handler

type Handler struct {
	log               *zerolog.Logger
	urlResolver       *URLResolver
	sess              Session
	processor         *Processor
	activityPubRouter chi.Router
	browserRouter     chi.Router
}

func NewHandler(log *zerolog.Logger, urlResolver *URLResolver, sess Session, processor *Processor) *Handler {
	h := &Handler{
		log:         log,
		urlResolver: urlResolver,
		sess:        sess,
		processor:   processor,
	}

	tracer := serverIOTracer{enable: true, log: log}

	fallback := chi.NewRouter()
	fallback.Use(sess.Middleware)
	fallback.Get("/", h.handleIndex)
	fallback.Get("/login", h.handleLoginGet)
	fallback.Post("/login", h.handleLoginPost)
	fallback.Post("/logout", h.handleLogoutPost)
	fallback.Get("/signup", h.handleSignupGet)
	fallback.Post("/signup", h.handleSignupPost)
	fallback.Get("/@{username}", h.handleUserGet)
	fallback.Post("/@{username}/follow", h.handleUserFollowPost)
	fallback.Post("/@{username}/unfollow", h.handleUserUnfollowPost)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer, tracer.middleware, chizerolog.LoggerMiddleware(log))
	router.Get("/.well-known/host-meta", h.handleWellKnownHostMetaGet)
	router.Get("/.well-known/nodeinfo", h.handleWellKnownNodeInfoGet)
	router.Get("/nodeinfo/2.1", h.handleNodeInfo2Dot1Get)
	router.Get("/.well-known/webfinger", h.handleWellKnownWebfingerGet)
	router.Get("/u/{accountID}", h.handleUGet)
	router.Get("/u/{accountID}/main-key", h.handleUserMainKeyGet)
	router.Get("/u/{accountID}/inbox", h.handleUserInboxGet)
	router.Post("/u/{accountID}/inbox", h.handleUserInboxPost)
	router.Post("/u/{accountID}/outbox", h.handleUserOutboxPost)
	router.Handle("/*", fallback)

	h.activityPubRouter = router
	h.browserRouter = fallback

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.activityPubRouter.ServeHTTP(w, r)
}

// GET /.well-known/host-meta
func (h *Handler) handleWellKnownHostMetaGet(w http.ResponseWriter, r *http.Request) {
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")

	w.Header().Set("Content-Type", "application/xrd+xml")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.Header))
	enc.Encode(internal.XMLHostMeta{
		XMLName: xml.Name{
			Local: "XRD",
		},
		Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
		Links: []internal.XMLHostMetaLink{
			{
				Rel:  "lrdd",
				Type: "application/xrd+xml",
				Template: fmt.Sprintf("%s/.well-known/webfinger?resource={uri}",
					h.urlResolver.myURLPrefix()),
			},
		},
	})
}

// GET /.well-known/nodeinfo
func (h *Handler) handleWellKnownNodeInfoGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(internal.JSONNodeInfo{
		Links: []internal.JSONNodeInfoLink{
			{
				Rel:  "http://nodeinfo.diaspora.software/ns/schema/2.1",
				Href: fmt.Sprintf("%s/nodeinfo/2.1", h.urlResolver.myURLPrefix()),
			},
		},
	})
}

// GET /nodeinfo/2.1
func (h *Handler) handleNodeInfo2Dot1Get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(internal.JSONNodeInfo2Dot1{
		Version: "2.1",
		Software: internal.JSONNodeInfo2Dot1Software{
			Name:    "activitypub",
			Version: "0.0.1",
		},
		Protocols: []string{
			"activitypub",
		},
		Services: internal.JSONNodeInfo2Dot1Services{
			Inbound:  []string{},
			Outbound: []string{},
		},
		OpenRegistrations: false,
		Usage:             internal.JSONNodeInfo2Dot1Usage{},
		Metadata:          internal.JSONNodeInfo2Dot1Metadata{},
	})
}

// Get /.well-known/webfinger
func (h *Handler) handleWellKnownWebfingerGet(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")

	actor, err := h.processor.Webfinger(r.Context(), resource)
	if err != nil {
		h.catchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/jrd+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(internal.JSONWebfinger{
		Subject: resource,
		Links: []internal.JSONWebfingerLink{
			{
				Rel:  "self",
				Type: "application/activity+json",
				Href: actor.ID,
			},
		},
	})
}

// GET /u/{userID}
func (h *Handler) handleUGet(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	actor, err := h.processor.GetLocalAccount(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.catchError(w, err)
	}

	w.Header().Set("Content-Type", "application/activity+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(actor)
}

// Get /u/{userID}/main-key
func (h *Handler) handleUserMainKeyGet(w http.ResponseWriter, r *http.Request) {
	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	mainKey, err := h.processor.GetMainKey(r.Context(), accountID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		h.catchError(w, err)
	}

	w.Header().Set("Content-Type", "application/activity+json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mainKey)
}

func (h *Handler) handleUserInboxGet(w http.ResponseWriter, r *http.Request) {
}

func (h *Handler) handleUserInboxPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.processor.ReceiveInbox(c, accountID, r.Body)
	if err != nil {
		h.catchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/activity+json")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleUserOutboxPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	accountID := chi.URLParam(r, "accountID")
	if accountID == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := h.processor.ReceiveOutbox(c, accountID, r.Body)
	if err != nil {
		h.catchError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/activity+json")
	w.WriteHeader(http.StatusOK)
}

// GET /
func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	_, ok := h.sess.Get(c, sessionKey).(string)
	if !ok {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<ul>
				<li><a href="/login">login</a></li>
				<li><a href="/signup">signup</a></li>
			</ul>
		`))
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<ul>
			<!--<li><a href="/users">users</a></li>-->
		</ul>
		<form method="POST" action="/logout">
			<button>logout</button>
		</form>
	`))
}

// GET /login
func (h *Handler) handleLoginGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<form method="POST">
			<input type="text" name="email" />
			<input type="password" name="password" />
			<input type="submit" />
		</form>
	`))
}

// POST /login
func (h *Handler) handleLoginPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	email := r.FormValue("email")
	password := r.FormValue("password")
	id, err := h.processor.Login(c, email, password)
	if err != nil {
		// TODO: bad request
		h.catchError(w, err)
		return
	}

	h.sess.Set(c, sessionKey, id)
	http.Redirect(w, r, "/", http.StatusFound)
}

// POST /logout
func (h *Handler) handleLogoutPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	h.sess.Clear(c)
	http.Redirect(w, r, "/", http.StatusFound)
}

// GET /signup
func (h *Handler) handleSignupGet(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`
		<form method="POST">
			<input type="text" name="email" />
			<input type="text" name="username" />
			<input type="password" name="password" />
			<input type="submit" />
		</form>
	`))
}

// POST /signup
func (h *Handler) handleSignupPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	email := r.FormValue("email")
	username := r.FormValue("username")
	password := r.FormValue("password")
	id, err := h.processor.Signup(c, email, username, password)
	if err != nil {
		h.catchError(w, err)
		return
	}

	h.sess.Set(c, sessionKey, id)
	http.Redirect(w, r, "/", http.StatusFound)
}

// GET /@{username}
func (h *Handler) handleUserGet(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	acct := chi.URLParam(r, "username")
	accountID, ok := h.sess.Get(c, sessionKey).(string)
	if !ok {
		accountID = ""
	}

	user, err := h.processor.View(c, accountID, acct)
	if err != nil {
		h.catchError(w, err)
		return
	}

	tmpl, _ := template.New("").Parse(`
		<h1>{{.username}}</h1>
		{{if .isFollow}}<p>フォロー中</p>{{end}}
		{{if .isFollower}}<p>フォロワー</p>{{end}}
		{{if .isFollow}}
		<form method="post" action="/@{{.acct}}/follow">
			<button type="submit">フォロー</button>
		</form>
		{{else}}
		<form method="post" action="/@{{.acct}}/unfollow">
			<button type="submit">解除フォロー</button>
		</form>
		{{end}}
		<p>{{.bio}}</p>
		<ul>
			{{range .links}}
				<li><a href="{{.link}}">{{.name}}</a></li>
			{{end}}
		</ul>
	`)

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	tmpl.Execute(w, map[string]interface{}{
		"acct":       acct,
		"username":   user.Actor.Username,
		"isFollow":   user.IsFollow,
		"isFollower": user.IsFollower,
	})
}

// POST /@{username}/follow
func (h *Handler) handleUserFollowPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	acct := chi.URLParam(r, "username")
	accountID, ok := h.sess.Get(c, sessionKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := h.processor.Follow(c, accountID, acct)
	if err != nil {
		h.catchError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/@%s", acct), http.StatusFound)
}

// POST /@{username}/unfollow
func (h *Handler) handleUserUnfollowPost(w http.ResponseWriter, r *http.Request) {
	c := r.Context()
	acct := chi.URLParam(r, "username")
	accountID, ok := h.sess.Get(c, sessionKey).(string)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err := h.processor.Unfollow(c, accountID, acct)
	if err != nil {
		h.catchError(w, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/@%s", acct), http.StatusFound)
}

func (h *Handler) catchError(w http.ResponseWriter, err error) {
	h.log.Error().Err(err).Send()
	http.Error(w, InternalServerError, http.StatusInternalServerError)
}

// urlResolver

func NewURLResolver(cfg *Config) *URLResolver {
	return &URLResolver{
		host:  cfg.Host,
		https: cfg.Https,
	}
}

type URLResolver struct {
	host  string
	https bool
}

func (u *URLResolver) myURLPrefix() string {
	scheme := "http"
	if u.https {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, u.host)
}

func (u *URLResolver) resolveActorURL(accountID string) string {
	return fmt.Sprintf("%s/u/%s", u.myURLPrefix(), accountID)
}

func (u *URLResolver) resolveMainKeyURL(accountID string) string {
	return fmt.Sprintf("%s/u/%s/main-key", u.myURLPrefix(), accountID)
}

func (u *URLResolver) resolveInboxURL(accountID string) string {
	return fmt.Sprintf("%s/u/%s/inbox", u.myURLPrefix(), accountID)
}

func (u *URLResolver) resolveOutboxURL(accountID string) string {
	return fmt.Sprintf("%s/u/%s/outbox", u.myURLPrefix(), accountID)
}

func (u *URLResolver) resolveActivityURL(activityID string) string {
	return fmt.Sprintf("%s/a/%s", u.myURLPrefix(), activityID)
}

func (u *URLResolver) getProfileURL(username string, host string) string {
	if host == "" || host == u.host {
		return fmt.Sprintf("%s/@%s", u.myURLPrefix(), username)
	}
	return fmt.Sprintf("%s/@%s@%s", u.myURLPrefix(), username, host)
}

// acct

type userAddr struct {
	preferredUsername string
	host              string
}

func parseAcctScheme(str string) (*userAddr, error) {
	prefix := "acct:"
	if !strings.HasPrefix(str, prefix) {
		return nil, fmt.Errorf("invalid acct: %s", str)
	}

	acctStr := strings.TrimPrefix(str, prefix)
	return parseUserAddr(acctStr)
}

func parseUserAddr(str string) (*userAddr, error) {
	acctStr := strings.TrimSuffix(str, "@")

	atIndex := strings.Index(acctStr, "@")
	if atIndex == -1 {
		return &userAddr{
			preferredUsername: acctStr,
		}, nil
	}

	return &userAddr{
		preferredUsername: acctStr[:atIndex],
		host:              acctStr[atIndex+1:],
	}, nil
}

type serverIOTracer struct {
	enable bool
	log    *zerolog.Logger
}

func (s *serverIOTracer) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.enable {
			body, _ := io.ReadAll(r.Body)
			r.Body.Close()
			br := bytes.NewReader(body)
			r.Body = io.NopCloser(br)

			header, _ := json.Marshal(r.Header)
			s.log.Trace().
				Str("path", r.URL.String()).
				RawJSON("header", header).
				Str("body", string(body)).
				Send()
		}

		next.ServeHTTP(w, r)
	})
}
