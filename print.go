package jz1

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func PrintStruct(obj interface{}) {
	b, _ := json.Marshal(obj)
	var out bytes.Buffer
	_ = json.Indent(&out, b, "", "    ")
	fmt.Println(out.String())
}
