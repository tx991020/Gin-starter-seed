package {package}

import (
	"fmt"

	dao {DaoPkgPath}
	"github.com/garyburd/redigo/redis"
	"strconv"
	"time"
	"errors"
)

var {table}Cache cache.Cache                // memory cache container
var {table}CacheLifespan time.Duration      // keys' life span in memory
var {table}CacheRedisKeyLifespan int        // keys' life span in redis 
var {table}PublishKey string                // publish key, dedup publish message

func {table}CacheInit(c *CacheConfig) {
	cache.Register(dao.{Table}Table, cache.NewMemoryCache)
	{table}CacheLifespan = time.Duration(c.Lifespan) * time.Second
	{table}CacheRedisKeyLifespan = c.RedisKeyLifespan
	l := new{Table}CacheLoader(c.Server, c.Password, c.Database, c.MaxIdle, c.MaxActive, c.IdleTimeout)
	{table}Cache, _ = cache.NewCache(dao.{Table}Table, fmt.Sprintf("{\"interval\":%d}", c.CachePurgeInterval), l)
	go {table}RedisSubscribe(l.pool)
	conn := l.pool.Get()
	defer conn.Close()

	if i, err := redis.Int(conn.Do("INCR", "PubKey")); err != nil {
		logs.Error("%v", err)
		panic(err)
	} else {
		{table}PublishKey = string(strconv.AppendInt(nil, int64(i), 10))
		logs.Info("PubKey: %v", {table}PublishKey)
	}
}

func {table}RedisSubscribe(pool *redis.Pool) {
	c := pool.Get()
	psc := redis.PubSubConn{Conn: c}
	psc.PSubscribe(dao.{Table}Table + "*")

	for {
		switch n := psc.Receive().(type) {
		case redis.Message:
			//       fmt.Printf("Message: %s %s\n", n.Channel, n.Data)
		case redis.PMessage:
			if n.Channel != dao.{Table}Table+":"+{table}PublishKey {
			//       fmt.Println("invalid: ", string(n.Data))
				{table}Cache.Invalid(string(n.Data))
			}
			//       fmt.Printf("PMessage: %s %s %s\n", n.Pattern, n.Channel, n.Data)
		case redis.Subscription:
			//       fmt.Printf("Subscription: %s %s %d\n", n.Kind, n.Channel, n.Count)
		case error:
			c.Close()
			logs.Error("error: %v\n", n)
			time.Sleep(1 * time.Second)
			c = pool.Get()
			psc = redis.PubSubConn{Conn: c}
			psc.PSubscribe(dao.{Table}Table + "*")
		}
	}
	c.Close()
}

type {table}CacheLoader struct {
	pool *redis.Pool
}

func new{Table}CacheLoader(server, password string, database, maxIdle, maxActive, idleTimeout int) *{table}CacheLoader {
  var dialOptions []redis.DialOption
  if len(password) > 0 {
    dialOptions = append(dialOptions, redis.DialPassword(password))
  }
  dialOptions = append(dialOptions, redis.DialDatabase(database))
	return &{table}CacheLoader{
		pool: &redis.Pool{
			MaxIdle:     maxIdle,
			MaxActive:   maxActive,
			IdleTimeout: time.Duration(idleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", server, dialOptions...)
				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

// key.expire or !key.exist then Load(key)
// 1. redis.get && expire redis ttl
// 2. return obj && expire memory ttl
func (l *{table}CacheLoader) Load(key string) (interface{}, time.Duration) {
	c := l.pool.Get()
	defer c.Close()

	var err error
	value := ""
	rk := dao.{Table}Table + ":" + key
	if value, err = redis.String(c.Do("GET", rk)); err != nil {
		//     logs.Error("%v, %v", err, rk)
		return nil, 0 // not exist is legal result
	} else {
		c.Do("EXPIRE", rk, {table}CacheRedisKeyLifespan)
	}

	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(key), &m, "json"); err != nil {
		logs.Error("%v, %v", err, key)
		return nil, 0
	}

	if m["_count"] != nil { // int
		count := 0
		if err := json.Unmarshal([]byte(value), &count, "json"); err != nil {
			logs.Error("%v, %v", err, value)
			return nil, 0
		}
		return count, time.Duration({table}CacheLifespan)
	} else if m["id"] != nil { // *dao.{Table}
		o := &dao.{Table}{}
		if err := json.Unmarshal([]byte(value), &o, "json"); err != nil {
			logs.Error("%v, %v", err, value)
			return nil, 0
		}
		return o, time.Duration({table}CacheLifespan)
	} else { // []int64
		i := []int64{}
		if err := json.Unmarshal([]byte(value), &i, "json"); err != nil {
			logs.Error("%v, %v", err, value)
			return nil, 0
		}
		return i, time.Duration({table}CacheLifespan)
	}
	return nil, 0
}

// cache.put && redis.set && publish or redis.get && cache.put
func (l *{table}CacheLoader) Put(key string, o interface{}) error {
	c := l.pool.Get()
	defer c.Close()

	rk := dao.{Table}Table + ":" + key
	value := string(hdutil.MustJson(o, "json", false))
	if _, err := redis.String(c.Do("SET", rk, value, "EX", {table}CacheRedisKeyLifespan)); err != nil { // ttl
		logs.Error("%v, %v", err, rk)
		return err
	}
	c.Do("PUBLISH", dao.{Table}Table+":"+{table}PublishKey, key)
	return nil
}

// redis.del after delete from database
func (l *{table}CacheLoader) Delete(key string) error {
	c := l.pool.Get()
	defer c.Close()

	rk := dao.{Table}Table + ":" + key
	c.Do("DEL", rk)
	c.Do("PUBLISH", dao.{Table}Table+":"+{table}PublishKey, key)
	return nil
}

func Set{Table}Cache(e *dao.{Table}) {
  if e == nil {
    return
  }
	k := string(hdutil.MustJson(map[string]interface{}{"id": e.Id}, "json", false))
  {table}Cache.Put(k, e, {table}CacheLifespan)
}

// db.create && cache.put
func Create{Table}(e *dao.{Table}) (*dao.{Table}, error) {
	o, err := dao.Create{Table}(e)
	if err != nil {
		return nil, err
	}
	if o != nil {
		k := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
		if o != nil {
			{table}Cache.Put(k, o, {table}CacheLifespan)
		}
	}
	return o, nil
}

func Invalid{Table}(id int64) {
	k := string(hdutil.MustJson(map[string]interface{}{"id": id}, "json", false))
  	{table}Cache.Delete(k)
}

// cache.get else (db.get && cache.put)
func Get{Table}(id int64) (*dao.{Table}, error) {
	k := string(hdutil.MustJson(map[string]interface{}{"id": id}, "json", false))
	cache, exist := {table}Cache.Get(k)
	if exist {
		if cache == nil {
			return nil, errors.New("item is empty")
		} else {
			return cache.(*dao.{Table}), nil
		}
	} else {
		o, err := dao.Get{Table}(id)
		if err != nil {
			return nil, err
		}
		{table}Cache.Put(k, o, {table}CacheLifespan)
		return o, nil
	}
}

func Invalid{Table}s(m map[string]interface{}) {
	if m != nil && m["id"] != nil {
		logs.Error("Invalid{Table}s with non nil attr: id")
    return
	}
	k := string(hdutil.MustJson(m, "json", false))
  {table}Cache.Delete(k)
}

// query attrs should not be with "id", use Get{Table} instead
// cache.get(query) && cache.get(id) else db.get(query) && cache.set(query, ids) && cache.set(id)
func Get{Table}s(m map[string]interface{}) ([]*dao.{Table}, error) {
	if m != nil && m["id"] != nil {
		logs.Error("Get{Table}s with non nil attr: id")
		return []*dao.{Table}{}, errors.New("invalid attr")
	}
	k := string(hdutil.MustJson(m, "json", false))
	cache, exist := {table}Cache.Get(k)
	if exist {
		swp := make([]*dao.{Table}, 0)
		for _, i := range cache.([]int64) {
			if item, err := Get{Table}(i); err == nil {
				swp = append(swp, item)
			}
		}
		return swp, nil
	} else {
		list, err := dao.Get{Table}s(m)
		if err != nil {
			return nil, err
		}
		ids := make([]int64, 0)
		for _, o := range list {
			ids = append(ids, o.Id)
			idk := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
			{table}Cache.Put(idk, o, {table}CacheLifespan)
		}
		{table}Cache.Put(k, ids, {table}CacheLifespan)
		return list, nil
	}
}

func Get{Table}sByIds(ids []int64) ([]*dao.{Table}, error) {
  m := map[string]interface{}{"_ids": ids}
	k := string(hdutil.MustJson(m, "json", false))
	cache, exist := {table}Cache.Get(k)
	if exist {
		swp := make([]*dao.{Table}, 0)
		for _, i := range cache.([]int64) {
			if item, err := Get{Table}(i); err == nil {
				swp = append(swp, item)
			}
		}
		return swp, nil
	} else {
		list, err := dao.Get{Table}sByIds(ids)
		if err != nil {
			return nil, err
		}
		for _, o := range list {
			idk := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
			{table}Cache.Put(idk, o, {table}CacheLifespan)
		}
		{table}Cache.Put(k, ids, {table}CacheLifespan)
		return list, nil
	}
}

// db.update && cache.put
func Update{Table}(id int64, m map[string]interface{}) (*dao.{Table}, int64, bool, error) {
	delete(m, "id")
	delete(m, "ctime")
	delete(m, "utime")
	k := string(hdutil.MustJson(map[string]interface{}{"id": id}, "json", false))
	o, aff, exist, err := dao.Update{Table}(id, m)
  if err == nil {
    {table}Cache.Put(k, o, {table}CacheLifespan)
  }
  return o, aff, exist, err
}

// db.del && cache.del
func Delete{Table}(id int64) error {
	k := string(hdutil.MustJson(map[string]interface{}{"id": id}, "json", false))
	err := dao.Delete{Table}(id)
	if err != nil {
		return err
	}
	if err := {table}Cache.Delete(k); err != nil {
		logs.Error("%v", err)
	}

	return nil
}

func Invalid{Table}sByTime(start time.Time, end time.Time, m map[string]interface{}, rawq string, limit int) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_start"] = hdutil.JSONTime{start}
	m["_end"] = hdutil.JSONTime{end}
	m["_limit"] = limit
	m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_start")
	delete(m, "_end")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")
	{table}Cache.Delete(k)
	{table}Cache.Delete(kc)
}

func Get{Table}sByTime(start time.Time, end time.Time, m map[string]interface{}, rawq string, limit int) ([]*dao.{Table}, int, error) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_start"] = hdutil.JSONTime{start}
	m["_end"] = hdutil.JSONTime{end}
	m["_limit"] = limit
	m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_start")
	delete(m, "_end")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")

	cache, exist := {table}Cache.Get(k)
	count, _ := {table}Cache.Get(kc)
	if exist {
		swp := make([]*dao.{Table}, 0)
		for _, id := range cache.([]int64) {
			if o, err := Get{Table}(id); err == nil {
				swp = append(swp, o)
			}
		}
		return swp, count.(int), nil
	} else {
		list, cnt, err := dao.Get{Table}sByTime(start, end, m, rawq, limit, true)
		if err != nil {
			return nil, 0, err
		}
		ids := make([]int64, 0)
		for _, o := range list {
			ids = append(ids, o.Id)
			idk := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
			{table}Cache.Put(idk, o, {table}CacheLifespan)
		}
		{table}Cache.Put(k, ids, {table}CacheLifespan)
		{table}Cache.Put(kc, cnt, {table}CacheLifespan)
		return list, cnt, nil
	}
}

func Invalid{Table}sById(id int64, m map[string]interface{}, rawq string, dir int, order int, ordcond string, limit int) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_id"] = id
	m["_order"] = order
	m["_ordcond"] = ordcond
	m["_limit"] = limit
	m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_id")
	delete(m, "_order")
	delete(m, "_ordcond")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")
	{table}Cache.Delete(k)
	{table}Cache.Delete(kc)
}

func Get{Table}sById(id int64, m map[string]interface{}, rawq string, dir int, order int, ordcond string, limit int) ([]*dao.{Table}, int, error) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_id"] = id
	m["_order"] = order
	m["_ordcond"] = ordcond
	m["_limit"] = limit
  m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_id")
	delete(m, "_order")
	delete(m, "_ordcond")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")

	cache, exist := {table}Cache.Get(k)
	count, _ := {table}Cache.Get(kc)
	if exist {
		swp := make([]*dao.{Table}, 0)
		for _, id := range cache.([]int64) {
			if o, err := Get{Table}(id); err == nil {
				swp = append(swp, o)
			}
		}
		return swp, count.(int), nil
	} else {
		list, cnt, err := dao.Get{Table}sById(id, m, rawq, dir, order, ordcond, limit, true)
		if err != nil {
			return nil, 0, err
		}
		ids := make([]int64, 0)
		for _, o := range list {
			ids = append(ids, o.Id)
			idk := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
			{table}Cache.Put(idk, o, {table}CacheLifespan)
		}
		{table}Cache.Put(k, ids, {table}CacheLifespan)
		{table}Cache.Put(kc, cnt, {table}CacheLifespan)
		return list, cnt, nil
	}
}

func Invalid{Table}sByOffset(offset int, m map[string]interface{}, rawq string, order int, ordcond string, limit int) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_offset"] = offset
	m["_order"] = order
	m["_ordcond"] = ordcond
	m["_limit"] = limit
  m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_offset")
	delete(m, "_order")
	delete(m, "_ordcond")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")
  {table}Cache.Delete(k)
  {table}Cache.Delete(kc)
}

func Get{Table}sByOffset(offset int, m map[string]interface{}, rawq string, order int, ordcond string, limit int) ([]*dao.{Table}, int, error) {
	if m == nil {
		m = map[string]interface{}{}
	}
	m["_offset"] = offset
	m["_order"] = order
	m["_ordcond"] = ordcond
	m["_limit"] = limit
	m["_rawq"] = rawq
	k := string(hdutil.MustJson(m, "json", false))
	m["_count"] = true
	kc := string(hdutil.MustJson(m, "json", false))
	delete(m, "_offset")
	delete(m, "_order")
	delete(m, "_ordcond")
	delete(m, "_limit")
	delete(m, "_rawq")
	delete(m, "_count")

	cache, exist := {table}Cache.Get(k)
	count, _ := {table}Cache.Get(kc)
	if exist {
		swp := make([]*dao.{Table}, 0)
		for _, id := range cache.([]int64) {
			if o, err := Get{Table}(id); err == nil {
				swp = append(swp, o)
			}
		}
		return swp, count.(int), nil
	} else {
		list, cnt, err := dao.Get{Table}sByOffset(offset, m, rawq, order, ordcond, limit, true)
		if err != nil {
			return nil, 0, err
		}
		ids := make([]int64, 0)
		for _, o := range list {
			ids = append(ids, o.Id)
			idk := string(hdutil.MustJson(map[string]interface{}{"id": o.Id}, "json", false))
			{table}Cache.Put(idk, o, {table}CacheLifespan)
		}
		{table}Cache.Put(k, ids, {table}CacheLifespan)
		{table}Cache.Put(kc, cnt, {table}CacheLifespan)
		return list, cnt, nil
	}
}
