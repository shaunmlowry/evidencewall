package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs request/response details in non-production environments.
// Sensitive fields are redacted from JSON bodies.
func RequestLogger(serviceName, environment string) gin.HandlerFunc {
	// Only enable detailed logs outside production
	if environment == "production" {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		start := time.Now()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		clientIP := c.ClientIP()

		var bodyPreview string
		ct := c.GetHeader("Content-Type")
		if strings.Contains(ct, "application/json") {
			// Read and restore body (limit to ~2MB)
			limited := io.LimitReader(c.Request.Body, 2*1024*1024)
			buf, _ := io.ReadAll(limited)
			c.Request.Body.Close()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))

			if len(buf) > 0 {
				var asAny interface{}
				if err := json.Unmarshal(buf, &asAny); err == nil {
					redactSensitive(asAny)
					if redacted, err2 := json.Marshal(asAny); err2 == nil {
						bodyPreview = string(redacted)
					}
				}
			}
		}

		c.Next()

		status := c.Writer.Status()
		latency := time.Since(start)
		if bodyPreview != "" {
			log.Printf("%s:req %s %s?%s ip=%s status=%d dur=%s body=%s", serviceName, method, path, query, clientIP, status, latency, bodyPreview)
		} else {
			log.Printf("%s:req %s %s?%s ip=%s status=%d dur=%s", serviceName, method, path, query, clientIP, status, latency)
		}
	}
}

// redactSensitive walks JSON-like structures and masks sensitive values.
func redactSensitive(v interface{}) {
	switch x := v.(type) {
	case map[string]interface{}:
		for k, val := range x {
			kl := strings.ToLower(k)
			if strings.Contains(kl, "password") || strings.Contains(kl, "secret") || strings.HasSuffix(kl, "token") {
				x[k] = "***REDACTED***"
				continue
			}
			redactSensitive(val)
		}
	case []interface{}:
		for i := range x {
			redactSensitive(x[i])
		}
	default:
		// scalars are fine
	}
}
