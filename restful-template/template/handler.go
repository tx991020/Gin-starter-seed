package handler

import (
	"gopkg.in/gin-gonic/gin.v1"
	"time"
	"github.com/gin-contrib/cache/persistence"


func HandlerInit(r gin.IRouter, pageCacheSvr, pageCachePwd string) {
	store = persistence.NewRedisCache(pageCacheSvr, pageCachePwd, time.Hour*24)
	storeInMem = persistence.NewInMemoryStore(time.Hour * 8)
	accessor = service.GetRedis(pageCacheSvr, pageCachePwd)

	{namelist}
}
