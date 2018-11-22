package plot

import (
	"html/template"

	"github.com/antongulenko/golib"
	"github.com/gin-gonic/gin"
)

// Requirement: go get github.com/mjibson/esc
//go:generate esc -o static_files_generated.go -pkg plotHttp -prefix static/ static

func (p *HttpPlotter) serve() error {
	engine := golib.NewGinEngine()
	index := template.New("index")
	indexStr, err := FSString(p.UseLocalStatic, "/index.html")
	if err != nil {
		return err
	}
	index.Parse(indexStr)
	engine.SetHTMLTemplate(index)

	engine.GET("/", p.serveMain)
	engine.GET("/metrics", p.serveListData)
	engine.GET("/data", p.serveData)
	engine.StaticFS("/static", FS(p.UseLocalStatic))

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
		c.JSON(200, p.allMetricData())
	} else {
		c.JSON(200, p.metricData(name))
	}
}
