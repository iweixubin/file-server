package web

import (
	"io"
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"file-server/internal"
)

var srcDirHandler http.Handler
var resizedDirHandler http.Handler

// 静态文件处理
func staticServer(w http.ResponseWriter, req *http.Request) {
	urlPath := strings.ToLower(req.URL.Path)
	if urlPath == "/favicon.ico" {
		return
	}

	//println(req.URL.Path)

	if !allowResize(urlPath) {
		srcDirHandler.ServeHTTP(w, req)
		return
	}

	tNow := time.Now()
	if internal.Cfg.CacheMode != internal.CacheModeDisk && internal.Cfg.HttpExpires != 0 {
		lastModified := req.Header.Get("If-Modified-Since")
		if lastModified != "" {
			if t, e := time.Parse(http.TimeFormat, lastModified); e != nil {
				t = t.Add(time.Minute * time.Duration(internal.Cfg.HttpExpires))
				if t.After(tNow) {
					w.WriteHeader(http.StatusNotModified)
					return
				}
			}
		}
	}

	if internal.Cfg.CacheMode == internal.CacheModeDisk {
		if _, ok := internal.CacheProxy.Get(urlPath); ok {
			resizedDirHandler.ServeHTTP(w, req)
			return
		}
	}

	bs, err := internal.Resize(urlPath)
	if err != nil {
		io.WriteString(w, err.Error())
		return
	}

	// 如果没有则说明生成调整大小
	if bs == nil || len(bs) == 0 {
		srcDirHandler.ServeHTTP(w, req)
		return
	}

	w.Header().Set("Content-Type", mime.TypeByExtension(path.Ext(urlPath)))
	w.Header().Set("Last-Modified", tNow.Format(http.TimeFormat))
	if internal.Cfg.CacheMode != internal.CacheModeDisk && internal.Cfg.HttpExpires != 0 {
		w.Header().Set("Expires", tNow.Add(time.Minute*time.Duration(internal.Cfg.HttpExpires)).Format(http.TimeFormat))
	}

	w.Write(bs)

	//imaging.Encode(w, tiny, f)

}

func RunHttp() {
	srcDirHandler = http.FileServer(http.Dir(internal.Cfg.SrcDir))

	if internal.Cfg.CacheMode == internal.CacheModeDisk {
		resizedDirHandler = http.FileServer(http.Dir(internal.Cfg.CacheModeDiskDir))
	}

	http.HandleFunc("/", staticServer)

	if err := http.ListenAndServe(internal.Cfg.Host, nil); err != nil {
		panic(err)
	}
}
