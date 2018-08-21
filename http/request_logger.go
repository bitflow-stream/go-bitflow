package plotHttp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type GinRequestLogger struct {
	Filename   string
	LogBody    bool
	LogHeaders bool
}

func LogGinRequests(filename string, logBody, logHeaders bool) gin.HandlerFunc {
	logger := GinRequestLogger{Filename: filename, LogBody: logBody, LogHeaders: logHeaders}
	return logger.LogRequest
}

func (l *GinRequestLogger) LogRequest(context *gin.Context) {
	defer context.Next()
	if l.Filename == "" {
		return
	}
	logData := l.formatRequest(context)
	err := l.appendToFile(l.Filename, logData)
	if err != nil {
		log.Errorf("Failed to write HTTP request log to %v: %v", l.Filename, err)
		log.Errorln("Data:", logData)
	}
}

func (l *GinRequestLogger) formatRequest(context *gin.Context) []byte {
	r := context.Request
	timeStr := time.Now().Format("2006-01-02 15:04:05.999")
	var result bytes.Buffer
	fmt.Fprintf(&result, "%v %v from %v: %v", timeStr, r.Method, context.ClientIP(), r.RequestURI)
	loggingHeaders := l.LogHeaders && len(r.Header) > 0
	if loggingHeaders {
		result.WriteString("\n")
		r.Header.Write(&result)
		result.Truncate(result.Len() - 4) // Delete the trailing "\r\n\r\n" characters
	}
	if l.LogBody && r.ContentLength > 0 {
		if !loggingHeaders {
			fmt.Fprintf(&result, "\nContent-Length: %v", r.ContentLength)
		}
		err := r.ParseForm()
		if err != nil {
			log.Errorln("Error reading and parsing request body:", err)
		}
		if r.PostForm != nil && len(r.PostForm) > 0 {
			result.WriteString("\n")
			result.WriteString(r.PostForm.Encode())
		} else {
			// Since the body was not parsed as a POST form, try to log the entire body
			var body io.ReadCloser
			if r.GetBody != nil {
				body, err = r.GetBody()
				if err != nil {
					log.Errorln("Error obtaining a copy of the request body:", err)
				}
			}
			if body == nil {
				body = r.Body
			}
			bodyData, err := ioutil.ReadAll(body)
			if err != nil {
				log.Errorln("Error reading request body: %v", err)
			} else if len(bodyData) > 0 {
				result.WriteString("\n")
				result.Write(bodyData)
			}
		}
	}
	result.WriteString("\n")
	return result.Bytes()
}

func (l *GinRequestLogger) appendToFile(filename string, data []byte) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
