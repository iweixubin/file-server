package main

import (
	"file-server/internal"
	"file-server/web"
)

func main() {
	println("FileServer Run On --→ " + internal.Cfg.Host)

	switch internal.Cfg.WebEngine {
	case "Http":
		web.RunHttp()
	case "FastHttp":
		web.RunFastHttp()
	default:
		panic("配置文件 WebEngine 的值应该是 Http 或 FastHttp")
	}
}
