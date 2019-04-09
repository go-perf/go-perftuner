package main

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func atoi(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

func marshalJSON(results interface{}) {
	raw, _ := json.MarshalIndent(results, "", "  ")
	fmt.Printf("%s", string(raw))
}
