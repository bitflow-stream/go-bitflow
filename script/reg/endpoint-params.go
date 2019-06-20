package reg

import (
	"fmt"
	"net/url"
)

func ParseEndpointUrl(urlStr string) (*url.URL, error) {
	urlStr = "http://" + urlStr // the url.Parse() routine requires a schema, which is stripped in bitflow.EndpointFactory.Create*
	return url.Parse(urlStr)
}

func ParseEndpointUrlParams(urlStr string, params RegisteredParameters) (*url.URL, map[string]interface{}, error) {
	parsedUrl, err := ParseEndpointUrl(urlStr)
	if err != nil {
		return nil, nil, err
	}
	parsedParams, err := ParseTypedQueryParameters(parsedUrl, params)
	return parsedUrl, parsedParams, err
}

func ParseQueryParameters(parsedUrl *url.URL) (map[string]string, error) {
	values, err := url.ParseQuery(parsedUrl.RawQuery)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for key, val := range values {
		if len(val) == 1 {
			result[key] = val[0]
		} else if len(val) > 1 {
			return nil, fmt.Errorf("Multiple values for URL query key '%v': %v", key, val)
		}
	}
	return result, nil
}

func ParseTypedQueryParameters(parsedUrl *url.URL, params RegisteredParameters) (map[string]interface{}, error) {
	paramMap, err := ParseQueryParameters(parsedUrl)
	if err != nil {
		return nil, err
	}
	return params.ParsePrimitives(paramMap)
}
