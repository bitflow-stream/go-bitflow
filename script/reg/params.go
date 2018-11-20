package reg

import (
	"fmt"
	"strconv"
	"time"
)

func ParameterError(name string, err error) error {
	return fmt.Errorf("Failed to parse '%v' parameter: %v", name, err)
}

func StrParam(params map[string]string, name string, defaultVal string, hasDefault bool, err *error) string {
	if *err != nil {
		return ""
	}
	strVal, ok := params[name]
	if ok {
		return strVal
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return ""
	}
}

func DurationParam(params map[string]string, name string, defaultVal time.Duration, hasDefault bool, err *error) time.Duration {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := time.ParseDuration(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return 0
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func IntParam(params map[string]string, name string, defaultVal int, hasDefault bool, err *error) int {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.Atoi(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return 0
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func FloatParam(params map[string]string, name string, defaultVal float64, hasDefault bool, err *error) float64 {
	if *err != nil {
		return 0
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.ParseFloat(strVal, 64)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return 0
	}
}

func BoolParam(params map[string]string, name string, defaultVal bool, hasDefault bool, err *error) bool {
	if *err != nil {
		return false
	}
	strVal, ok := params[name]
	if ok {
		parsed, parseErr := strconv.ParseBool(strVal)
		if parseErr != nil {
			*err = ParameterError(name, parseErr)
			return false
		}
		return parsed
	} else if hasDefault {
		return defaultVal
	} else {
		*err = ParameterError(name, fmt.Errorf("Missing required parmeter"))
		return false
	}
}
