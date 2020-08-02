package web

import (
	"mime"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"file-server/internal"
)

var srcDirFastHandler fasthttp.RequestHandler
var resizedDirFastHandler fasthttp.RequestHandler

func fastHandler(ctx *fasthttp.RequestCtx) {
	urlPath := strings.ToLower(string(ctx.RequestURI()))
	if urlPath == "/favicon.ico" {
		return
	}

	if !allowResize(urlPath) {
		srcDirFastHandler(ctx)
		return
	}

	tNow := time.Now()
	if internal.Cfg.CacheMode != internal.CacheModeDisk && internal.Cfg.HttpExpires != 0 {
		lastModified := string(ctx.Request.Header.Peek("If-Modified-Since"))
		if lastModified != "" {
			if t, e := time.Parse(http.TimeFormat, lastModified); e == nil {
				t = t.Add(time.Minute * time.Duration(internal.Cfg.HttpExpires))
				if t.After(tNow) {
					ctx.SetStatusCode(http.StatusNotModified)
					return
				}
			}
		}
	}

	if internal.Cfg.CacheMode == internal.CacheModeDisk {
		if _, ok := internal.CacheProxy.Get(urlPath); ok {
			resizedDirFastHandler(ctx)
			return
		}
	}

	bs, err := internal.Resize(urlPath)
	if err != nil {
		return
	}

	// 如果没有则说明生成调整大小
	if bs == nil || len(bs) == 0 {
		srcDirFastHandler(ctx)
		return
	}

	ctx.Response.Header.Set("Content-Type", mime.TypeByExtension(path.Ext(urlPath)))
	ctx.Response.Header.Set("Last-Modified", tNow.Format(http.TimeFormat))
	if internal.Cfg.CacheMode != internal.CacheModeDisk && internal.Cfg.HttpExpires != 0 {
		ctx.Response.Header.Set("Expires", tNow.Add(time.Minute*time.Duration(internal.Cfg.HttpExpires)).Format(http.TimeFormat))
	}

	ctx.SetBody(bs)
}

func RunFastHttp() {
	srcDirFastHandler = (&fasthttp.FS{Root: internal.Cfg.SrcDir}).NewRequestHandler()

	if internal.Cfg.CacheMode == internal.CacheModeDisk {
		resizedDirFastHandler = (&fasthttp.FS{Root: internal.Cfg.CacheModeDiskDir}).NewRequestHandler()
	}

	if e := fasthttp.ListenAndServe(internal.Cfg.Host, fastHandler); e != nil {
		panic(e)
	}
}
