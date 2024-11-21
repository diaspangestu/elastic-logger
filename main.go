package elasticLogger

import (
	"bytes"
	"fmt"
	"github.com/diaspangestu/elastic-logger/elastic"
	"github.com/gin-gonic/gin"
	"io"
	"time"
)

type ResponseCapture struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *ResponseCapture) Write(b []byte) (int, error) {
	r.body.Write(b) // capture response body
	return r.ResponseWriter.Write(b)
}

func (l *elastic.Logger) Middleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// start timer
		start := time.Now()

		// capture req body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		responseBodyBuffer := &bytes.Buffer{}
		writer := &ResponseCapture{
			ResponseWriter: c.Writer,
			body:           responseBodyBuffer,
		}
		c.Writer = writer

		// process request
		c.Next()

		// capture response body and status
		status := c.Writer.Status()
		responseBody := responseBodyBuffer.String()
		headers := formatHeaders(c.Request.Header)

		logMsg := elastic.LogMessage{
			ServiceName: l.service,
			Hostname:    c.Request.Host,
			Timestamp:   time.Now().Format("20060201"),
			HttpStatus:  fmt.Sprintf("%d", status),
			Header:      headers,
			Request:     requestBody,
			Response:    string(responseBody),
		}

		// log to elastic
		err := l.Log(logMsg)
		if err != nil {
			c.Error(err)
		}

		// log processing line
		duration := time.Since(start)
		c.Header("X-Processing-Time", duration.String())
	}
}

func formatHeaders(headers map[string][]string) string {
	var formatted string
	for key, value := range headers {
		formatted = fmt.Sprintf("%s: %s\n", key, value)
	}
	return formatted
}
