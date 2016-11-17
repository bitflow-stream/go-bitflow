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

func init() {
	flag.BoolVar(&useLocalStatic, "local", useLocalStatic, "Use local static files, instead of the ones embedded in the binary")
}

func GoServe(endpoint string) {
	go func() {
		if err := Serve(endpoint); err != nil {
			log.Fatalln(err)
		}
	}()
}

func Serve(endpoint string) error {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	logHandler := ginrus.Ginrus(log.StandardLogger(), log.DefaultTimestampFormat, false)
	engine.Use(logHandler, gin_recover)

	index := template.New("index")
	index.Parse(FSMustString(useLocalStatic, "/index.html"))
	engine.SetHTMLTemplate(index)

	engine.GET("/", serveMain)
	engine.GET("/metrics", serveListData)
	engine.GET("/data", serveData)
	engine.StaticFS("/static", FS(useLocalStatic))
	return engine.Run(endpoint)
}

func serveMain(c *gin.Context) {
	c.HTML(200, "index", nil)
}

func serveListData(c *gin.Context) {
	c.JSON(200, HttpPlotter.metricNames())
}

func serveData(c *gin.Context) {
	name := c.Request.FormValue("metric")
	if len(name) == 0 {
		c.Writer.WriteString("Need 'metrics' parameter")
		c.Writer.WriteHeader(http.StatusBadRequest)
		return
	}
	c.JSON(200, HttpPlotter.metricData(name))
}

func gin_recover(c *gin.Context) {
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
