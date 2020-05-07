package main

import (
	"fmt"
	"go.opencensus.io/plugin/ochttp"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"
	"net/http"
)

func main() {

	ocagentHost := "oc-collector.tracing:55678"
	oce, _ := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(1*time.Second),
		ocagent.WithAddress(ocagentHost),
		ocagent.WithServiceName("guesttracker"))

	trace.RegisterExporter(oce)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	r := gin.Default()
	r.POST("/track-guest", func(c *gin.Context) {
		fmt.Println(c.Request.Header)
		fmt.Println(c.Request.Host)
		fmt.Println("Tracking the user")
		context := c.Request.Context()
		span := trace.FromContext(context)
		defer span.End()
		span.Annotate([]trace.Attribute{trace.StringAttribute("annotated", "guesttrackervalue")}, "guesttracker annotation check")
		span.AddAttributes(trace.StringAttribute("span-add-attribute", "guesttrackervalue"))
		var json LoginData
		if err := c.ShouldBindJSON(&json); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"Added": json})
	})
	//r.Run(":8081") // listen and serve on 0.0.0.0:8081
	http.ListenAndServe( // nolint: errcheck
		"0.0.0.0:8081",
		&ochttp.Handler{
			Handler: r,
			GetStartOptions: func(r *http.Request) trace.StartOptions {
				startOptions := trace.StartOptions{}

				if r.URL.Path == "/metrics" {
					startOptions.Sampler = trace.NeverSample()
				}

				return startOptions
			},
		},)
}



type LoginData struct {
	UserName     string `json:"username"`
	Email string `json:"email"`
}

