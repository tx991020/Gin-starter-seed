package {package}

import (
	"reflect"
	"net/http"
	"gopkg.in/gin-gonic/gin.v1"
)

func {Table}Init(r gin.IRouter) {
	rt := r.Group(`/{table}s`)
	rt.POST(
		``,
		xutil.GRequestBodyObject(reflect.TypeOf(dao.{Table}{}), "json"),
		Create{Table},
	)

	rt.PUT(
		`/:{table}Id`,
		xutil.GPathRequireInt("{table}Id"),
		xutil.GRequestBodyMap,
		Update{Table},
	)
	rt.GET(
		`/:{table}Id`,
		xutil.GPathRequireInt("{table}Id"),
		Get{Table},
	)

	r.GET(
		`/{table}sByFilter`,
		xutil.GQueryOptionalStringDefault("filter", "{}"),
		xutil.GQueryOptionalIntDefault("current", 1),
		xutil.GQueryOptionalIntDefault("pageSize", 10),
		Get{Table}ByFilter,
	)
	rt.DELETE(
		`/:{table}Id`,
		xutil.GPathRequireInt("{table}Id"),
		Delete{Table},
	)
}
// @Produce  json
// @Param product body dao.{Table} true "Create {Table}"
// @Success 200 {object} dao.{Table}
// @Router /v1/{table}s [post]
func Create{Table}(c *gin.Context) {
	u := c.MustGet("requestBody").(*dao.{Table})
	item := cache.Create{Table}(u)

	service.InvalidCache{Table}ByFilter()

	c.Data(http.StatusOK, "", xutil.GR(service.NoneError, item))
}
// @Produce  json
// @Param	id			path 	int	true		"The id you want to update"
// @Param	body		body 	dao.Product	true		"content"
// @Success 200 {object} dao.{Table}
// @router /v1/{{table}s/{id} [put]
func Update{Table}(c *gin.Context) {
	m := c.MustGet("requestBody").(map[string]interface{})
	delete(m, "id")
	r, _, _, err := cache.Update{Table}(c.MustGet("{table}Id").(int64), xutil.MCamelToSnake(m))
	c.Data(http.StatusOK, "", xutil.GR(service.ErrorDatabase(err), r))
}
// @Produce  json
// @Param	id		path 	int	true		"id"
// @Success 200 {object} dao.{Table}
// @router /v1/{table}s/{id} [get]
func Get{Table}(c *gin.Context) {
	r := cache.Get{Table}(c.MustGet("{table}Id").(int64))
	if r == nil {
		c.Data(http.StatusOK, "", xutil.GR(service.NoneError, nil))
		return
	}
	c.Data(http.StatusOK, "", xutil.GR(service.NoneError, r))
}

// @Produce  json
// @Param	id		path 	int	true		"id"
// @Success 200 {object} dao.{Table}
// @router /v1/{table}s/{id} [get]
func Delete{Table}(c *gin.Context) {
	cache.Delete{Table}(c.MustGet("{table}Id").(int64))
	service.InvalidCache{Table}ByFilter()
	c.Data(http.StatusOK, "", xutil.GR(service.NoneError, nil))
}

// @Produce  json
// @Param  filter  query string true "{}"
// @Param  pageSize  query int false "PageSize"
// @Param  current query   int false "Current"
// @Success 200 {object} dao.{Table}
// @Router /v1/{table}ByFilter [get]
func Get{Table}ByFilter(c *gin.Context) {
	current := c.MustGet("current").(int64)
	pagesize := c.MustGet("pageSize").(int64)

	jsonstring := c.MustGet("filter").(string)
	filter := make(map[string]interface{})
	err := json.Unmarshal([]byte(jsonstring), &filter)
	if err != nil {
		c.Data(http.StatusOK, "", xutil.GR(service.ErrorEntityInvalid, nil))
		return
	}

	rels, cnt := service.Fetch{Table}ByFilter(xutil.MCamelToSnake(filter), current, pagesize)

	c.Header("X-total-count", xutil.Itoa(int64(cnt)))
	c.Data(http.StatusOK, "", xutil.GR(service.NoneError, rels))
}
