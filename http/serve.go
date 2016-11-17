package plotHttp

import (
	"flag"
	"html/template"
	"net/http"
	"net/http/httputil"

	log "github.com/Sirupsen/logrus"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
)

var useLocalStatic = false
var ginLogHandler = ginrus.Ginrus(log.StandardLogger(), log.DefaultTimestampFormat, false)

func init() {
	flag.BoolVar(&useLocalStatic, "http-local-static", useLocalStatic,
		"For -e http: serve local static files, instead of the ones embedded in the binary")

	gin.SetMode(gin.ReleaseMode)
}

func (p *HttpPlotter) serve() error {
	engine := gin.New()
	engine.Use(ginLogHandler, ginRecover)

	index := template.New("index")
	index.Parse(FSMustString(useLocalStatic, "/index.html"))
	engine.SetHTMLTemplate(index)

	engine.GET("/", p.serveMain)
	engine.GET("/metrics", p.serveListData)
	engine.GET("/data", p.serveData)
	engine.StaticFS("/static", FS(useLocalStatic))

	return engine.Run(p.Endpoint)
}

func (p *HttpPlotter) serveMain(c *gin.Context) {
	c.HTML(200, "index", nil)
}

func (p *HttpPlotter) serveListData(c *gin.Context) {
	c.JSON(200, p.metricNames())
}

func (p *HttpPlotter) serveData(c *gin.Context) {
	name := c.Request.FormValue("metric")
	if len(name) == 0 {
		c.Writer.WriteString("Need 'metrics' parameter")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	c.JSON(200, p.metricData(name))
}

func ginRecover(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			stack := stack(3)
			httprequest, _ := httputil.DumpRequest(c.Request, false)
			log.Errorf("[Recovery] panic recovered:\n%s\n%s\n%s", string(httprequest), err, stack)
			c.AbortWithStatus(500)
		}
	}()
	c.Next()
}
