package internal

import (
	"errors"
	"path"
	"regexp"
	"strconv"
	"strings"
)

var sizeRegexp *regexp.Regexp

func initSizeRegexp() {
	var err error
	sizeRegexp, err = regexp.Compile(Cfg.SizeRegexp)
	if err != nil {
		panic(err)
	}
}

func getSize(urlPath string) (srcPath string, w, h int, err error) {
	ss := sizeRegexp.FindAllStringSubmatch(urlPath, -1)
	if ss == nil {
		return
	}
	if len(ss[0]) != 3 {
		return
	}

	if w, err = strconv.Atoi(ss[0][1]); err != nil {
		return
	}

	if h, err = strconv.Atoi(ss[0][2]); err != nil {
		return
	}

	if len(Cfg.SizeLimits) != 0 {
		pass := false
		for _, v := range Cfg.SizeLimits {
			if v.Width == w && v.Height == h {
				pass = true
				break
			}
		}
		if !pass {
			err = errors.New("宽高不在指定范围中~")
			return
		}
	}

	if w == 0 || h == 0 {
		return
	}

	index := strings.Index(urlPath, ss[0][0])

	srcPath = urlPath[0:index]

	if path.Ext(urlPath) != path.Ext(srcPath) {
		err = errors.New("后缀名不一致~")
	}

	return
}
