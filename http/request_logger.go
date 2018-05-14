package plotHttp

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type GinRequestLogger struct {
	Filename string
	LogBody  bool
}

func LogGinRequests(filename string, logBody bool) gin.HandlerFunc {
	logger := GinRequestLogger{Filename: filename, LogBody: logBody}
	return logger.LogRequest
}

func (l *GinRequestLogger) LogRequest(context *gin.Context) {
	defer context.Next()
	if l.Filename == "" {
		return
	}
	logData := l.requestToString(context)
	err := l.appendToFile(l.Filename, logData+"\n")
	if err != nil {
		log.Errorf("Failed to write HTTP request log to %v: %v", l.Filename, err)
		log.Errorln("Data:", logData)
	}
}

func (l *GinRequestLogger) requestToString(context *gin.Context) string {
	r := context.Request
	timeStr := time.Now().Format("2006-01-02 15:04:05.999")
	result := fmt.Sprintf("%v %v from %v: %v", timeStr, r.Method, context.ClientIP(), r.RequestURI)
	if l.LogBody {
		if r.ContentLength > 0 {
			result += fmt.Sprintf("\nContent Length: %v", r.ContentLength)
		}
		bodyData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorln("Error reading request body: %v", err)
		} else if len(bodyData) > 0 {
			result += "\n" + string(bodyData)
		} else if len(r.PostForm) > 0 {
			// If the body is empty, the POST form was probably already parsed
			result += "\n" + r.PostForm.Encode()
		}
	}
	return result
}

func (l *GinRequestLogger) appendToFile(filename string, data string) error {
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0664)
	if err != nil {
		return err
	}
	n, err := f.WriteString(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
