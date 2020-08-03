package web

import (
	"file-server/internal"
	"mime"
	"strings"
)

func init() {
	if mime.TypeByExtension(".bmp") == "" {
		mime.AddExtensionType(".bmp", "image/bmp")
	}
}

func allowResize(urlPath string) bool {
	for _, v := range internal.Cfg.SupportExtResize {
		if strings.HasSuffix(urlPath, v) {
			return true
		}
	}
	return false
}
