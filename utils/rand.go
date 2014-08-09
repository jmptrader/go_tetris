package utils

import "github.com/astaxie/beego/utils"

var RandString = func(n int) string { return string(utils.RandomCreateBytes(n)) }
