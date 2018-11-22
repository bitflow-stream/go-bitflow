package plugin

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/bitflow-stream/go-bitflow/bitflow"
	log "github.com/sirupsen/logrus"
)

func ParseEndpointUrl(urlStr string) (string, string, map[string]string, error) {
	parsedUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", "", nil, fmt.Errorf("Failed to parse plugin URL: %v", err)
	}
	urlParams := parsedUrl.Query()
	params := make(map[string]string, len(urlParams))
	for key, value := range urlParams {
		paramVal := ""
		if len(value) == 1 {
			paramVal = value[0]
		} else if len(value) > 1 {
			return "", "", nil, fmt.Errorf("Multiple values for URL query key '%v': %v", key, value)
		}
		params[key] = paramVal
	}
	return parsedUrl.Host, parsedUrl.Path, params, nil
}

func ParseQueryParameters(queryStr string) (map[string]string, error) {
	values, err := url.ParseQuery(queryStr)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for key, val := range values {
		result[key] = strings.Join(val, " ")
	}
	return result, nil
}

func LogPluginDataSource(p BitflowPlugin, sourceName bitflow.EndpointType) {
	log.Debugf("Plugin %v: Registering data source '%v'", p.Name(), sourceName)
}

func LogPluginProcessor(p BitflowPlugin, stepName string) {
	log.Debugf("Plugin %v: Registering processing step '%v'", p.Name(), stepName)
}
