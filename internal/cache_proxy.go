package internal

import (
	"context"
	"os"
	"path"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/patrickmn/go-cache"
)

type CacheMode int

const (
	CacheModeNone CacheMode = iota
	CacheModeDisk
	CacheModeMemory
	CacheModeRedis
)

var CacheProxy ICache

func initCache() {
	switch Cfg.CacheMode {
	case CacheModeDisk:
		CacheProxy = &dCache{dicFileExist: make(map[string]bool)}
	case CacheModeMemory:
		CacheProxy = &mCache{theCache: cache.New(time.Minute, time.Minute)}
	case CacheModeRedis:
		if opt, e := redis.ParseURL(Cfg.CacheModeRedisConn); e != nil {
			panic(e)
		} else {
			_rCache := &rCache{}
			_rCache.ctx = context.Background()
			_rCache.theCache = redis.NewClient(opt)
			if e := _rCache.theCache.Ping(_rCache.ctx).Err(); e != nil {
				panic(e)
			}
			CacheProxy = _rCache
		}
	}

}

// ICache
type ICache interface {
	Get(key string) ([]byte, bool)
	Set(key string, img []byte)
}

type dCache struct {
	// 记录是否存在硬盘中
	dicFileExist map[string]bool

	// 确保并发的时候只向硬盘生产一次缩略图
	onceCheck sync.Map

	// 记录已经存在的目录
	dirMaked sync.Map
}

func (c *dCache) isFileExist(p string) bool {
	f, err := os.Stat(p)
	if f == nil || os.IsNotExist(err) {
		return false
	}

	return true
}

func (c *dCache) Get(key string) ([]byte, bool) {
	if Cfg.CacheMode != CacheModeDisk {
		return nil, false
	}

	// 在内存中判断，当然是快过硬盘很多倍~
	if v, ok := c.dicFileExist[key]; ok {
		return nil, v
	}

	exist := c.isFileExist(path.Join(Cfg.CacheModeDiskDir, key))
	if exist { // 减少对 dicHardDriver 的操作，已减少并发的竞态问题
		c.dicFileExist[key] = exist
	}

	return nil, exist
}

func (c *dCache) Set(key string, img []byte) {
	if Cfg.CacheMode != CacheModeDisk {
		return
	}

	fullPath := path.Join(Cfg.CacheModeDiskDir, key)

	exist := c.isFileExist(fullPath)
	if exist {
		c.dicFileExist[key] = true
		return
	}

	// 假设并发多个相同的url
	// http://127.0.0.1:8080/img/name.png_180x180.png 第一次才生成
	// http://127.0.0.1:8080/img/name.png_180x180.png 其它的不执行生成了
	if _, doing := c.onceCheck.LoadOrStore(key, nil); !doing {
		// 将 urlPath 存 onceCheck，执行完才移除
		// 那么在执行期间，其它相同url并发就不会执行 go func()...
		go func() {
			// 再判断一次是否存在
			exist := c.isFileExist(fullPath)
			if exist {
				c.dicFileExist[key] = true
				return
			}

			dir, _ := path.Split(fullPath)

			// 判断目录是否已经存在
			if _, dirExist := c.dirMaked.LoadOrStore(dir, nil); !dirExist {
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					c.dirMaked.Delete(dir)
				}
			}

			//
			f, err := os.Create(fullPath)
			created := true
			if err != nil {
				created = false
			}

			_, err = f.Write(img)
			if err != nil {
				created = false
				os.Remove(fullPath)
			}

			if created { // 减少对 dicHardDriver 的操作，以减少并发的竞态问题
				c.dicFileExist[key] = created
			}

			c.onceCheck.Delete(key)
		}()
	}
}

// 内存缓存
type mCache struct {
	theCache *cache.Cache
}

func (c *mCache) Get(key string) ([]byte, bool) {
	if v, exist := c.theCache.Get(key); exist {
		if bs, ok := v.([]byte); ok {
			return bs, true
		}
	}

	return nil, false
}

func (c *mCache) Set(key string, img []byte) {
	c.theCache.Set(key, img, time.Second*time.Duration(Cfg.CacheModeMemoryExpires))
}

// redis 缓存
type rCache struct {
	ctx      context.Context
	theCache *redis.Client
}

func (c *rCache) Get(key string) ([]byte, bool) {
	if bs, e := c.theCache.Get(c.ctx, key).Bytes(); e == nil {
		return bs, len(bs) != 0
	}

	return nil, false
}

func (c *rCache) Set(key string, img []byte) {
	c.theCache.Set(c.ctx, key, img, time.Second*time.Duration(Cfg.CacheModeRedisExpires))
}
