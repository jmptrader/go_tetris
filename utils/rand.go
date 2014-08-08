// 2013 Author: Beego
package utils

import "crypto/rand"

const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func RandString(n int) string {
	if n <= 0 {
		return ""
	}
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

func getRand() string {
	return RandString(16)
}
