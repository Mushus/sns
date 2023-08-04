package main

import (
	"encoding/json"
	"fmt"
)

type ContextProps struct {
	Context json.RawMessage `json:"@context"`
}

type JSONActivity struct {
	ContextProps
	Type string          `json:"type"`
	Hoge json.RawMessage `json:"hoge,omitempty"`
}

func main() {

	var activity JSONActivity
	err := json.Unmarshal([]byte(`{
		"@context": "https://www.w3.org/ns/activitystreams",
		"type": "Follow",
		"actor": "https://m.mushus.net/users/mus_rt",
		"id": "https://m.mushus.net/01VDXNRGG0181AVW9Q3221SR6C",
		"object": {
			"actor": "https://m.mushus.net/users/mus_rt",
			"id": "https://m.mushus.net/users/mus_rt/follow/01SN9MBGSWR97ZAJC82Z0YZRXF",
			"object": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf",
			"to": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf",
			"type": "Follow"
		},
		"to": "https://test.mushus.net/u/aaa9c72b1ce847e5b14e5427710140bf"
	}`), &activity)

	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v", activity)

	b, err := json.Marshal(activity)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
