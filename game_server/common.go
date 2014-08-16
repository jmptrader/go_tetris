package main

import (
	"encoding/json"
	"fmt"
)

// response from game server to client
type responseData struct {
	Description string      `json:"desc"`
	Data        interface{} `json:"data"`
}

func newResponse(desc string, data interface{}) responseData {
	return responseData{Description: desc, Data: data}
}

func (r responseData) String() string {
	return fmt.Sprintf("\ndescription: %v, data: %v", r.Description, r.Data)
}

func (r responseData) toJson() string {
	b, err := json.Marshal(r)
	if err != nil {
		log.Debug("can not json marshal the data %v: %v", r, err)
	}
	return string(b)
}
