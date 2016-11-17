package plotHttp

import (
	"flag"
	"net/http/httputil"

	"github.com/antongulenko/go-bitflow"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"

	log "github.com/Sirupsen/logrus"
)

var useLocalStatic = false

func init() {
	flag.BoolVar(&useLocalStatic, "local", useLocalStatic, "Use local static files, instead of the ones embedded in the binary")
}

func Serve(endpoint string) error {
	logHandler := ginrus.Ginrus(log.StandardLogger(), log.DefaultTimestampFormat, false)

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(logHandler, gin_recover)
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	engine.StaticFS("static", FS(useLocalStatic))
	return engine.Run(endpoint)
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

func NewHttpPlotter() bitflow.SampleProcessor {
	return new(bitflow.AbstractProcessor)
}
