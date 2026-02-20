package plan

import (
	"encoding/json"
	"fmt"
)

type Change struct {
	Scope   string      `json:"scope"`
	Target  string      `json:"target"`
	Action  string      `json:"action"`
	Details interface{} `json:"details"`
}

type Plan struct {
	Changes  []Change `json:"changes"`
	Warnings []string `json:"warnings"`
}

func Print(p Plan) {
	b, _ := json.MarshalIndent(p, "", "  ")
	fmt.Println(string(b))
}
