package utils

import "strings"

func GetIp(ctx interface{}) string {
	ip := strings.Split(getContext(ctx).Request.RemoteAddr, ":")[0]
	if ip != "127.0.0.1" {
		return ip
	}
	if realIp := getContext(ctx).Request.Header.Get("X-Real-Ip"); realIp != "" {
		return realIp
	}
	if realIp := getContext(ctx).Request.Header.Get("X-Forwaded-For"); realIp != "" {
		return realIp
	}
	return ""
}
