package activitypub

import (
	"encoding/json"
	"encoding/xml"
	"errors"
)

// findActorIDFromWebfinger - WebfingerからActorIDを取得する
// 見つからない場合は空文字を返す
func findActorIDFromWebfinger(webfinger *JSONWebfinger) string {
	for _, link := range webfinger.Links {
		if link.Rel == "self" && link.Type == "application/activity+json" {
			return link.Href
		}
	}
	return ""
}

type XMLHostMeta struct {
	XMLName xml.Name          `xml:"XRD"`
	Links   []XMLHostMetaLink `xml:"Link"`
	Xmlns   string            `xml:"xmlns,attr"`
}

type XMLHostMetaLink struct {
	Rel      string `xml:"rel,attr"`
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}

type JSONNodeInfo struct {
	Links []JSONNodeInfoLink `json:"links"`
}

type JSONNodeInfoLink struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type JSONNodeInfo2Dot1 struct {
	OpenRegistrations bool                      `json:"openRegistrations"`
	Protocols         []string                  `json:"protocols"`
	Software          JSONNodeInfo2Dot1Software `json:"software"`
	Usage             JSONNodeInfo2Dot1Usage    `json:"usage"`
	Services          JSONNodeInfo2Dot1Services `json:"services"`
	Metadata          JSONNodeInfo2Dot1Metadata `json:"metadata"`
	Version           string                    `json:"version"`
}

type JSONNodeInfo2Dot1Software struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type JSONNodeInfo2Dot1Usage struct {
	Users JSONNodeInfo2Dot1UsageUsers `json:"users"`
}

type JSONNodeInfo2Dot1UsageUsers struct {
	Total int `json:"total"`
}

type JSONNodeInfo2Dot1Services struct {
	Inbound  []string `json:"inbound"`
	Outbound []string `json:"outbound"`
}

type JSONNodeInfo2Dot1Metadata struct {
	MaxNoteLength int `json:"maxNoteLength"`
}

type JSONWebfinger struct {
	Subject string              `json:"subject"`
	Links   []JSONWebfingerLink `json:"links"`
}

type JSONWebfingerLink struct {
	Rel  string `json:"rel"`
	Type string `json:"type"`
	Href string `json:"href"`
}

type JSONActor struct {
	Context                   json.RawMessage       `json:"@context"`
	ID                        string                `json:"id,omitempty"`
	Type                      string                `json:"type,omitempty"`
	Inbox                     string                `json:"inbox,omitempty"`
	Outbox                    string                `json:"outbox,omitempty"`
	Name                      string                `json:"name,omitempty"`
	PreferredUsername         string                `json:"preferredUsername"`
	Summary                   string                `json:"summary,omitempty"`
	URL                       string                `json:"url,omitempty"`
	Icon                      JSONActorIcon         `json:"icon,omitempty"`
	PublicKey                 JSONPublicKey         `json:"publicKey,omitempty"`
	Followers                 string                `json:"followers,omitempty"`
	Following                 string                `json:"following,omitempty"`
	Featured                  string                `json:"featured,omitempty"`
	Attachment                []JSONActorAttachment `json:"attachment,omitempty"`
	ManuallyApprovesFollowers bool                  `json:"manuallyApprovesFollowers,omitempty"`
	Discoverable              bool                  `json:"discoverable,omitempty"`
}

type JSONActorIcon struct {
	Type      string `json:"type"`
	MediaType string `json:"mediaType"`
	URL       string `json:"url"`
}

type JSONPublicKey struct {
	ID           string `json:"id"`
	Owner        string `json:"owner"`
	PublicKeyPem string `json:"publicKeyPem"`
}

type JSONActorAttachment struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type JSONMainKey struct {
	Context           json.RawMessage `json:"@context"`
	ID                string          `json:"id"`
	Type              string          `json:"type"`
	PreferredUsername string          `json:"preferredUsername"`
	PublicKey         JSONPublicKey   `json:"publicKey"`
}

type JSONOrderedCollection struct {
	Context      json.RawMessage `json:"@context"`
	Summary      string          `json:"summary"`
	ID           string          `json:"id"`
	Type         string          `json:"type"`
	TotalItems   int             `json:"totalItems"`
	OrderedItems []string        `json:"orderedItems"`
}

type JSONActivityType struct {
	Type string `json:"type"`
}

type JSONObject struct {
	Context []string `json:"@context,omitempty"`
	// ttachment
	// attributedTo
	// audience
	// content
	// context
	// name
	// endTime
	// generator
	// icon
	// image
	// inReplyTo
	// location
	// preview
	// published
	// replies
	// startTime
	// summary
	// tag
	// updated
	// url
	// to
	// bto
	// cc
	// bcc
	// mediaType
	// duration
}

type JSONLDFollow struct {
	Context json.RawMessage `json:"@context,omitempty"`
	JSONFollow
}

type JSONFollow struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Actor  string          `json:"actor"`
	Object json.RawMessage `json:"object"`
	To     json.RawMessage `json:"to,omitempty"`
}

func (f *JSONFollow) ParseTo() (string, error) {
	if f.To == nil {
		return "", errors.New("to is nil")
	}

	if f.To[0] == '[' {
		var to []string
		if err := json.Unmarshal(f.To, &to); err != nil {
			return "", err
		}
		return to[0], nil
	}

	var to string
	if err := json.Unmarshal(f.To, &to); err != nil {
		return "", err
	}

	return to, nil
}

type JSONLDAccept struct {
	Context json.RawMessage `json:"@context,omitempty"`
	JSONAccept
}

type JSONAccept struct {
	Type   string          `json:"type"`
	ID     string          `json:"id"`
	Actor  string          `json:"actor"`
	Object json.RawMessage `json:"object"`
}

//	{
//	    "@context": "https://www.w3.org/ns/activitystreams",
//	    "actor": "https://m.mushus.net/users/mus_rt",
//	    "id": "https://m.mushus.net/01VDXNRGG0181AVW9Q3221SR6C",
//	    "object": {
//	        "actor": "https://m.mushus.net/users/mus_rt",
//	        "id": "https://m.mushus.net/users/mus_rt/follow/01SN9MBGSWR97ZAJC82Z0YZRXF",
//	        "object": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf",
//	        "to": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf",
//	        "type": "Follow"
//	    },
//	    "to": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf",
//	    "type": "Undo"
//	}
type JSONUndo struct {
	Context json.RawMessage `json:"@context,omitempty"`
	Type    string          `json:"type"`
	ID      string          `json:"id"`
	Actor   string          `json:"actor"`
	Object  json.RawMessage `json:"object"`
	To      json.RawMessage `json:"to,omitempty"`
}

type JSONActivityIn struct {
	Type   string          `json:"type,omitempty"`
	Actor  string          `json:"actor,omitempty"`
	Object json.RawMessage `json:"object,omitempty"`
	raw    json.RawMessage
}

func (a *JSONActivityIn) UnmarshalJSON(b []byte) error {
	a.raw = b
	return json.Unmarshal(b, a)
}
