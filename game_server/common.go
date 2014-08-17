package main

import "fmt"

// response from game server to client
type responseData struct {
	Desc string      `json:"desc"`
	Data interface{} `json:"data"`
}

func newResponse(desc string, data interface{}) responseData {
	return responseData{Desc: desc, Data: data}
}

func (r responseData) String() string {
	return fmt.Sprintf("\ndescription: %v, data: %v", r.Desc, r.Data)
}

func (r responseData) toJson() map[string]interface{} {
	return map[string]interface{}{
		"desc": r.Desc,
		"data": r.Data,
	}
}

// func (r responseData) toJson() string {
// 	b, err := json.Marshal(r)
// 	if err != nil {
// 		log.Debug("can not json marshal the data %v: %v", r, err)
// 	}
// 	return string(b)
// }
