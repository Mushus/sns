package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/kelseyhightower/envconfig"
)

const AppConfigPrefix = "myapp"

func main() {
	if err := startServer(); err != nil {
		log.Println(err)
	}
}

type serverConfig struct {
	isDev   bool   `envconfig:"dev" default:"false"`
	host    string `envconfig:"host" default:"localhost:8080"`
	port    int    `envconfig:"port" default:"8080"`
	isHttps bool   `envconfig:"https" default:"true"`
}

func startServer() error {
	var cfg serverConfig
	if err := envconfig.Process(AppConfigPrefix, &cfg); err != nil {
		return fmt.Errorf("server configuration error: %w", err)
	}

	h := newHandler(&cfg)
	app := fiber.New()
	app.Use(logger.New(logger.Config{
		Format: "${pid} ${locals:requestid} ${status} - ${method} | ${reqHeaders} | ${path} ${queryParams}​\n",
	}))
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})
	app.Get("/.well-known/host-meta", h.getWellKnownHostMeta)
	app.Get("/.well-known/nodeinfo", h.getNodeInfo)
	app.Get("/nodeinfo/2.1", h.getNodeInfo2dot1)
	app.Get("/.well-known/webfinger", h.getWebFinger)
	app.Get("/@:username", h.getUsername)
	app.Get("/@:username/inbox", h.getUserInbox)
	app.Post("/@:username/inbox", h.postUserInbox)
	app.Get("/@:username/outbox", h.getUserOutbox)
	return app.Listen(fmt.Sprintf(":%d", cfg.port))
}

func newHandler(cfg *serverConfig) *handler {
	baseUrl := func() string {
		scheme := "http"
		if cfg.isHttps {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s", scheme, cfg.host)
	}
	return &handler{
		host:    cfg.host,
		baseUrl: baseUrl(),
	}
}

type handler struct {
	host    string
	baseUrl string
}

func (h *handler) getWellKnownHostMeta(c *fiber.Ctx) error {
	prefix := h.baseUrl
	body := fmt.Sprintf(`<?xml version="1.0"?>
<XRD xmlns="http://docs.oasis-open.org/ns/xri/xrd-1.0">
<Link rel="lrdd" type="application/xrd+xml" template="%s/.well-known/webfinger?resource={uri}" />
</XRD>
`, prefix)
	c.Set("Content-Type", "application/xrd+xml")
	return c.SendString(body)
}
func (h *handler) getNodeInfo(c *fiber.Ctx) error {
	prefix := h.baseUrl
	json := map[string]any{
		"links": []any{
			map[string]any{
				"rel":  "http://nodeinfo.diaspora.software/ns/schema/2.1",
				"href": prefix + "/nodeinfo/2.1",
			},
		},
	}
	return c.JSON(json)
}

func (h *handler) getNodeInfo2dot1(c *fiber.Ctx) error {
	json := map[string]any{
		"openRegistrations": false,
		"protocols": []string{
			"activitypub",
		},
		"software": map[string]any{
			"name":    "sns",
			"version": "0.1.0",
		},
		"usage": map[string]any{
			"users": map[string]any{
				"total": 1,
			},
		},
		"services": map[string]any{
			"inbound":  []any{},
			"outbound": []any{},
		},
		"metadata": map[string]any{},
		"version":  "2.1",
	}
	return c.JSON(json)
}

func (h *handler) getWebFinger(c *fiber.Ctx) error {
	resource := c.Query("resource")
	prefix := "acct:"
	suffix := "@" + h.host

	if !strings.HasPrefix(resource, prefix) || !strings.HasSuffix(resource, suffix) {
		return c.SendStatus(fiber.StatusBadRequest)
	}

	account := resource[len(prefix) : len(resource)-len(suffix)]
	if account != "dummy" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	actorHref := fmt.Sprintf("%s/@%s", h.baseUrl, account)
	json := map[string]any{
		"subject": resource,
		"aliases": []string{
			actorHref,
		},
		"links": []any{
			map[string]any{
				"rel":  "http://webfinger.net/rel/profile-page",
				"type": "text/html",
				"href": actorHref,
			},
			map[string]any{
				"rel":  "self",
				"type": "application/activity+json",
				"href": actorHref,
			},
		},
	}

	c.Set("Content-Type", "application/jrd+json; charset=utf-8")
	return c.JSON(json)
}

func (h *handler) getUsername(c *fiber.Ctx) error {
	username := c.Params("username")
	if username != "dummy" {
		return c.SendStatus(fiber.StatusNotFound)
	}
	prefix := h.baseUrl
	actorUrl := fmt.Sprintf("%s/@%s", prefix, username)
	publicKeyId := fmt.Sprintf("%s/#main-key", actorUrl)
	userInbox := fmt.Sprintf("%s/@%s/inbox", prefix, username)
	userOutbox := fmt.Sprintf("%s/@%s/outbox", prefix, username)
	json := map[string]any{
		"@context": []string{
			"https://www.w3.org/ns/activitystreams",
			"https://w3id.org/security/v1",
		},
		"id":                actorUrl,
		"type":              "Person",
		"preferredUsername": username,
		"inbox":             userInbox,
		"outbox":            userOutbox,
		"discoverable":      true,
		"publicKey": map[string]any{
			"id":           publicKeyId,
			"owner":        actorUrl,
			"publicKeyPem": "dummy",
		},
	}
	return c.JSON(json)
}

func (h *handler) getUserInbox(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *handler) postUserInbox(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}

func (h *handler) getUserOutbox(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNotImplemented)
}
