package {package}

import ()

type CacheConfig struct {
	Server             string   `json:"server"`                 // redis address
	Password           string   `json:"password"`               // redis password
  	Database           int      `json:"database"`               // redis database
	MaxIdle            int      `json:"maxIdle"`                // redis max idle connections
	MaxActive          int      `json:"maxActive"`              // redis max active connections
	IdleTimeout        int      `json:"idleTimeout"`            // redis idle connection timeout, in seconds
	RedisKeyLifespan   int      `json:"redisKeyLifespan"`       // redis key lifespan, in seconds
	CachePurgeInterval int      `json:"cachePurgeInterval"`     // memory cache purge interval, in seconds
	Lifespan           int      `json:"lifespan"`               // memory cache k-v lifespan, in seconds
	Tables             []string `json:"tables"`                 // tables affected by this config
}

var cacheInitFuncs map[string]func(c *CacheConfig) = map[string]func(c *CacheConfig){
  {namelist}
}

type CacheModuleConfig struct {
  CC []*CacheConfig `json:"cacheConfig"`
}

func CacheInit(conf *CacheModuleConfig) {
	for _, f := range cacheInitFuncs {
		f(conf.CC[0])
	}
}
