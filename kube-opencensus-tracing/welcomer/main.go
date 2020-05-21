package main

import (
	"bytes"
	"contrib.go.opencensus.io/exporter/ocagent"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/trace"
	"io/ioutil"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

var (
	guestrackerhost = os.Getenv("GUEST_TRACKER_HOST")
)

func main() {
	if guestrackerhost == "" {
		guestrackerhost = "localhost:8081"
	}
	fmt.Println("GUEST_TRACKER_HOST =", guestrackerhost)

	ocagentHost := "localhost:55678"
	oce, _ := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(1*time.Second),
		ocagent.WithAddress(ocagentHost),
		ocagent.WithServiceName("guesttracker"))

	trace.RegisterExporter(oce)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	r := gin.Default()

	r.GET("/welcome", func(c *gin.Context) {
		//_, span := trace.StartSpan(c, "/welcome")
		// http_server_route=/welcome tag is set
		context := c.Request.Context()
		ochttp.SetRoute(context, "/welcome")

		span := trace.FromContext(context)
		defer span.End()
		span.Annotate(
			[]trace.Attribute{
				trace.StringAttribute("callType", "Prateek1"),
			}, "annotation check 1")

		callGuestTracker(c, span)
		c.JSON(200, gin.H{
			"message": "Hello Folks .. You are welcome(Shhh... and also tracked by guesttracker)!!",
		})
	})
	//r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	http.ListenAndServe( // nolint: errcheck
		"0.0.0.0:8080",
		&ochttp.Handler{
			Handler: r,
			GetStartOptions: func(r *http.Request) trace.StartOptions {
				startOptions := trace.StartOptions{}

				if r.URL.Path == "/metrics" {
					startOptions.Sampler = trace.NeverSample()
				}

				return startOptions
			},
		})
}

func callGuestTracker(c *gin.Context, span *trace.Span) {
	reqBody, err := json.Marshal(map[string]string{
		"username": "Bruce Wayne",
		"email":    "batman@loreans.com",
	})
	if err != nil {
		print(err)
	}
	client := &http.Client{Transport: &ochttp.Transport{Base: &http.Transport{}}}
	context := c.Request.Context()
	//fmt.Println("ginContext: ", c)
	//fmt.Println("requestContext: ", context)

	time.Sleep(time.Millisecond * 125)

	req, _ := http.NewRequestWithContext(context, "POST", "http://"+guestrackerhost+"/track-guest", bytes.NewBuffer(reqBody))
	ctx, startSpan := trace.StartSpan(context, "abcd")
	clientTrace := ochttp.NewSpanAnnotatingClientTrace(req, startSpan)
	ctx = httptrace.WithClientTrace(ctx, clientTrace)
	req = req.WithContext(ctx)
	trace.FromContext(ctx).Annotate([]trace.Attribute{
		trace.StringAttribute("request-body", req.Host+req.URL.RequestURI()),
	}, string(reqBody))

	//startSpan.Annotate(
	//	[]trace.Attribute{
	//		trace.StringAttribute("callType", "Prateek2"),
	//	}, "welcomervalue-->guesttracker annotation check")
	//startSpan.AddAttributes(trace.StringAttribute("span-add-attribute", "welcomervalue"))

	resp, err := client.Do(req)

	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	respString := string(body)
	fmt.Println(respString)
	trace.FromContext(ctx).Annotate([]trace.Attribute{
		trace.StringAttribute("downstream-response", "http://"+guestrackerhost+"/track-guest"),
	}, respString)
	startSpan.End()

}
