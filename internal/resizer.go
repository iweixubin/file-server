package internal

import (
	"bytes"
	"path"

	"github.com/disintegration/imaging"
)

func Resize(urlPath string) ([]byte, error) {
	if Cfg.CacheMode == CacheModeMemory || Cfg.CacheMode == CacheModeRedis {
		if v, ok := CacheProxy.Get(urlPath); ok {
			return v, nil
		}
	}

	ext := path.Ext(urlPath)

	f, err := imaging.FormatFromExtension(ext)
	if err != nil {
		return nil, err
	}

	p, width, height, err := getSize(urlPath)
	if err != nil {
		return nil, err
	}

	if width == 0 || height == 0 {
		return nil, nil
	}

	srcImg, err := imaging.Open(path.Join(Cfg.SrcDir, p))
	if err != nil {
		return nil, err
	}

	tiny := imaging.Resize(srcImg, width, height, imaging.CatmullRom)

	buffer := &bytes.Buffer{}
	imaging.Encode(buffer, tiny, f)

	bs := buffer.Bytes()

	CacheProxy.Set(urlPath, bs)

	return bs, nil
}
