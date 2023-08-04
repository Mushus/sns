package activitypub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Mushus/activitypub/internal"
	"github.com/Mushus/activitypub/lib/crypt"
	"github.com/go-fed/httpsig"
)

type RemoteServer struct {
	softwareName string
	cli          *http.Client
	urlResolver  *URLResolver
}

func NewRemoteServer(cfg *Config, urlResolver *URLResolver) *RemoteServer {
	cli := &http.Client{}

	return &RemoteServer{
		softwareName: cfg.SoftwareName,
		cli:          cli,
		urlResolver:  urlResolver,
	}
}

func (s *RemoteServer) GetMainKey(c context.Context, mainKeyID string) (*internal.JSONPublicKey, error) {
	res, err := s.cli.Get(mainKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get main key: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get main key: status code %d", res.StatusCode)
	}

	var mainKey internal.JSONPublicKey
	if err := json.NewDecoder(res.Body).Decode(&mainKey); err != nil {
		return nil, fmt.Errorf("failed to decode main key: %w", err)
	}

	return &mainKey, nil
}

func (s *RemoteServer) GetWebfinger(c context.Context, host string, resource string) (*internal.JSONWebfinger, error) {
	uri := url.URL{
		Scheme: "https",
		Host:   host,
		Path:   ".well-known/webfinger",
		RawQuery: url.Values{
			"resource": []string{resource},
		}.Encode(),
	}
	res, err := s.cli.Get(uri.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get webfinger: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get webfinger: status code %d", res.StatusCode)
	}

	var webfinger internal.JSONWebfinger
	if err := json.NewDecoder(res.Body).Decode(&webfinger); err != nil {
		return nil, fmt.Errorf("failed to decode webfinger: %w", err)
	}

	return &webfinger, nil
}

func (s *RemoteServer) GetActor(c context.Context, account *Account, actorID string) (*internal.JSONActor, error) {
	req, err := http.NewRequestWithContext(c, http.MethodGet, actorID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/activity+json")

	res, err := s.GetSecureRequest(account, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get actor: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get actor: status code %d", res.StatusCode)
	}

	var actor internal.JSONActor
	if err := json.NewDecoder(res.Body).Decode(&actor); err != nil {
		return nil, fmt.Errorf("failed to decode actor: %w", err)
	}

	return &actor, nil
}

func (s *RemoteServer) GetInbox(c context.Context, account *Account, inboxURL string) (*internal.JSONOrderedCollection, error) {
	req, err := http.NewRequestWithContext(c, http.MethodGet, inboxURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	res, err := s.GetSecureRequest(account, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get inbox: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get inbox: status code %d", res.StatusCode)
	}

	var inbox internal.JSONOrderedCollection
	if err := json.NewDecoder(res.Body).Decode(&inbox); err != nil {
		return nil, fmt.Errorf("failed to decode inbox: %w", err)
	}

	return &inbox, nil
}

func (s *RemoteServer) PostInbox(c context.Context, account *Account, inboxURL string, bodyJSON any) error {
	body, err := json.Marshal(bodyJSON)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(c, http.MethodPost, inboxURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/activity+json")

	res, err := s.PostSecureRequest(account, req, body)
	if err != nil {
		return fmt.Errorf("failed to post inbox: %w", err)
	}

	defer func() {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post inbox: status code %d", res.StatusCode)
	}

	return nil
}

func (s *RemoteServer) GetSecureRequest(account *Account, req *http.Request) (*http.Response, error) {
	now := time.Now()
	// HTTP の Date ヘッダーは現在時間ではなくGTM
	req.Header.Set("Date", now.UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Set("Host", req.Host)
	req.Header.Set("User-Agent", s.softwareName)

	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256, httpsig.RSA_SHA512}
	digestAlgorithm := httpsig.DigestSha256
	headersToSign := []string{httpsig.RequestTarget, "host", "date", "user-agent"}

	signer, _, err := httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	privateKey, err := crypt.ConvertPrivateKey(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKeyURL := s.urlResolver.resolveMainKeyURL(account.ID)
	if err := signer.SignRequest(privateKey, publicKeyURL, req, nil); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return s.cli.Do(req)
}

func (s *RemoteServer) PostSecureRequest(account *Account, req *http.Request, body []byte) (*http.Response, error) {
	now := time.Now()
	// HTTP の Date ヘッダーは現在時間ではなくGTM
	req.Header.Set("Date", now.UTC().Format("Mon, 02 Jan 2006 15:04:05")+" GMT")
	req.Header.Set("Host", req.Host)
	req.Header.Set("User-Agent", s.softwareName)

	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256, httpsig.RSA_SHA512}
	digestAlgorithm := httpsig.DigestSha256
	headersToSign := []string{httpsig.RequestTarget, "host", "date", "user-agent", "digest"}

	signer, _, err := httpsig.NewSigner(prefs, digestAlgorithm, headersToSign, httpsig.Signature, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	privateKey, err := crypt.ConvertPrivateKey(account.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKeyURL := s.urlResolver.resolveMainKeyURL(account.ID)
	if err := signer.SignRequest(privateKey, publicKeyURL, req, body); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return s.cli.Do(req)
}
