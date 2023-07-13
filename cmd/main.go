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
	IsDev   bool   `envconfig:"dev" default:"false"`
	Host    string `envconfig:"host" default:"localhost:8080"`
	Port    int    `envconfig:"port" default:"8080"`
	IsHttps bool   `envconfig:"https" default:"true"`
}

func startServer() error {
	var cfg serverConfig
	if err := envconfig.Process(AppConfigPrefix, &cfg); err != nil {
		return fmt.Errorf("server configuration error: %w", err)
	}
	fmt.Printf("%#v\n", cfg)
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
	return app.Listen(fmt.Sprintf(":%d", cfg.Port))
}

func newHandler(cfg *serverConfig) *handler {
	baseUrl := func() string {
		scheme := "http"
		if cfg.IsHttps {
			scheme = "https"
		}
		return fmt.Sprintf("%s://%s", scheme, cfg.Host)
	}
	return &handler{
		host:    cfg.Host,
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
	// "-----BEGIN RSA PRIVATE KEY-----\nMIIEpQIBAAKCAQEA51j2f9I+IRlMj/IKlkCoXfLHRl9/19B+rDQsvq61lE8qla1y\n+/UHtDfkmu42X3xTQDVAFrOilEBXHO4XfRcI6JXOPCwxRMNEDEzuDLbtdtaKAx78\nPplqSWanbYoRtSG8JNcOu+EKwfVAZ+mBMUsMliplaL1Vig8aLQzYetrCgBnztBDb\nd1pijqKDvrTZLmJKv9H+LFBoVJ0b0w+QqnuXArXfPf8PaVVQpWnRXRbiOz5Wh4w2\ntHDAGc80wAkNa9TQIZbsfClt6/aBbcO0L2wmTvfCMnEt1IbuNBTIMc0EBLM1KUfR\nRt3X+hIQ7fRCo+VjWuVFnDAgBDhnznaXu8vkrQIDAQABAoIBAQC62uL1qJUf3LQ3\nC1K7uuuCPiXv1BCI+lBlvBprxObKLAsEK7zUfjtDt2VAMajfBKalFJ13+I0W2sTB\njBiSozlBykVx2mvM0z4yBSy8Pj+cHXoJPUyVLwpm0K/oTH0y5FV3F/BBlWk/8Vuc\n0j/T1X8MBqAzscDWKo6E1nw+9lPkbAhQyYXoYmDc+2Y7UOBFg6s9SfXDk8PBsQbD\nR80szfP68f1WJhJVnN+YanMCAPw+Vt8YrMJVdM7pUkEl+w+KGZ/9DQsfs54t67p+\n1YRVq/80azRkOD3gyPM9ShxpQi4DiXoVJWOPAUYVL6umTddl1MtP6jxbbvB57lAu\n+K3LLwwNAoGBAPuKxPXvUdAteIabaOudOIM895J3zMPV6ZM2f4iU1tfdKxvGZBG6\nJxAsMz6Q+3iaosCOK9FGpGZmXZKk7em82cEOYb24FbyeKL+CNDF49kFzhdaGqBNJ\nHA0l651BsYYAKFF7NKNKAK7K0yg6S0iPNJ1nf4LY1YXgS05PRKl5H2ljAoGBAOty\nknM3VWPzBVLjC7kmOSGGTO83UvB7D4SWFsOQ0tP3qQInmfGobtqq/PnRaZwNzlrw\niBoqUkt7SezG+j7EGtj0GvbYUW7zu56N7MSo3nDgNix/5xkUPf5Hqao2jhFaK/ba\nWBpsSc2bYIHwrqHH/F+mXffbjaeRCL+JEQzBBd6vAoGAP973rj8LdiHlpcBWfuVY\nETLs5jsXOm7ZtXC0J3krqHpXVOEmTb4H5zph9LQZtoEFbIFtLOGUIxBBGFhatOwo\nGrZNKUBR/KfoTuB/4kQFu47a4CMnEGaTAd+sGS0yJ4Vot2/iaMgErl2AConqzczX\nHlTGcvIeHbVbSdIk7Cd+S2MCgYEApXiZGmZaGeuS40T0WURWxIvph/m+zYn/RvRg\nvUMMGLKm0f/Y/nCcsAuZzUzyxx0g2OLRFGqH+cqFEuZouzIBmFY+mRtAaBTd2Dnw\nm+n+ox/Akxe05/hE9W+R+zFqOSHBYjTj1HYkjF7VvZzUbpjpcqOuyOJBtPGGT25a\nUDdcE7sCgYEAm0fqc9HKvHaA86D3BRFWUV7PHP5JxURLPcCKaNVNOuQ923pjcP14\nbJJB9zQdKn0UPqKAzSnW3mfIXXT4vYoAvaktCNJi6QPFM51FQLXFZaCmmH8VL/ho\nptgSdpj4xVOGSdFdYKbwmziBn+NI31ie42G6VdBgRePbeHhZPs7uy6s=\n-----END RSA PRIVATE KEY-----\n"

	pbulicKey := "-----BEGIN RSA PUBLIC KEY-----\nMIIBCgKCAQEA51j2f9I+IRlMj/IKlkCoXfLHRl9/19B+rDQsvq61lE8qla1y+/UH\ntDfkmu42X3xTQDVAFrOilEBXHO4XfRcI6JXOPCwxRMNEDEzuDLbtdtaKAx78Pplq\nSWanbYoRtSG8JNcOu+EKwfVAZ+mBMUsMliplaL1Vig8aLQzYetrCgBnztBDbd1pi\njqKDvrTZLmJKv9H+LFBoVJ0b0w+QqnuXArXfPf8PaVVQpWnRXRbiOz5Wh4w2tHDA\nGc80wAkNa9TQIZbsfClt6/aBbcO0L2wmTvfCMnEt1IbuNBTIMc0EBLM1KUfRRt3X\n+hIQ7fRCo+VjWuVFnDAgBDhnznaXu8vkrQIDAQAB\n-----END RSA PUBLIC KEY-----\n"
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
			"publicKeyPem": pbulicKey,
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
