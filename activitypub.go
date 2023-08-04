package activitypub

import "github.com/Mushus/activitypub/internal"

// findActorIDFromWebfinger - WebfingerからActorIDを取得する
// 見つからない場合は空文字を返す
func findActorIDFromWebfinger(webfinger *internal.JSONWebfinger) string {
	for _, link := range webfinger.Links {
		if link.Rel == "self" && link.Type == "application/activity+json" {
			return link.Href
		}
	}
	return ""
}
