package internal

import (
	"bytes"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
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

	imgBuffer := &bytes.Buffer{}

	if ext == ".gif" {
		// https://github.com/disintegration/imaging/issues/23
		// https://zhuanlan.zhihu.com/p/27718135
		// 还是不行
		gifFile, err := os.Open(path.Join(Cfg.SrcDir, p))
		if err != nil {
			return nil, err
		}
		defer gifFile.Close()

		g, err := gif.DecodeAll(gifFile)
		if err != nil {
			return nil, err
		}

		tGif := &gif.GIF{}
		tGif.Config = image.Config{
			ColorModel: g.Config.ColorModel,
			Width:      width,
			Height:     height,
		}
		tGif.BackgroundIndex = g.BackgroundIndex
		tGif.Delay = g.Delay
		tGif.Disposal = g.Disposal
		tGif.LoopCount = g.LoopCount

		firstFrame := g.Image[0].Bounds()
		b := image.Rect(0, 0, firstFrame.Dx(), firstFrame.Dy())
		im := image.NewRGBA(b)

		for i, frame := range g.Image {
			bounds := frame.Bounds()
			draw.Draw(im, bounds, frame, bounds.Min, draw.Over)
			t := imaging.Thumbnail(im, width, height, imaging.NearestNeighbor)
			g.Image[i] = imageToPaletted(t)
		}

		g.Config.Width = width
		g.Config.Height = height

		gif.EncodeAll(imgBuffer, g)

	} else {
		srcImg, err := imaging.Open(path.Join(Cfg.SrcDir, p))
		if err != nil {
			return nil, err
		}

		tiny := imaging.Thumbnail(srcImg, width, height, imaging.CatmullRom)

		imaging.Encode(imgBuffer, tiny, f)
	}

	buffer := imgBuffer.Bytes()

	CacheProxy.Set(urlPath, buffer)

	return buffer, nil
}

func imageToPaletted(img image.Image) *image.Paletted {
	b := img.Bounds()
	pm := image.NewPaletted(b, palette.Plan9)
	draw.FloydSteinberg.Draw(pm, b, img, image.ZP)
	return pm
}
