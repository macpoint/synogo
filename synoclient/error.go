package synoclient

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SynoError ...
type CommonSynoError struct {
	code   int
	reason string
}

// GenericError ...
type GenericError struct {
	desc string
}

type ApplicationError struct {
	code   int
	reason string
}

var commonSynoErrors = map[int]string{
	100: "Unknown error",
	101: "Invalid parameter",
	102: "The requested API does not exist",
	103: "The requested method does not exist",
	104: "The requested version does not support the functionality",
	105: "The logged in session does not have permission",
	106: "Session timeout",
	107: "Session interrupted by duplicate login",
}

func (synoerror *ApplicationError) Error() string {
	return fmt.Sprintf("Application error: (%v) %v", synoerror.code, synoerror.reason)
}

func HandleCommonSynoError(responseData interface{}) error {
	errorBlock := responseData.(map[string]interface{})["error"]
	// error -> code
	errorCode := int(errorBlock.(map[string]interface{})["code"].(float64))

	// check if we are handling common Syno errors (100-107)
	if _, ok := commonSynoErrors[errorCode]; ok {
		return &CommonSynoError{code: errorCode}
	}

	// this error should be handled in individual services
	return errors.New("error")
}

func HandleApplicationError(response string, err error, errorCodes map[int]string) error {
	switch err.(type) {
	default:
		return getAppError(response, errorCodes)
	case *CommonSynoError:
		return err
	case *GenericError:
		return err
	}
}

func getAppError(response string, errorCodes map[int]string) error {
	var responseData interface{}
	json.Unmarshal([]byte(response), &responseData)
	errorBlock := responseData.(map[string]interface{})["error"]
	// error -> code
	errorCode := int(errorBlock.(map[string]interface{})["code"].(float64))
	r := errorCodes[errorCode]
	return &ApplicationError{code: errorCode, reason: r}
}

func (synoerror *CommonSynoError) Error() string {
	synoerror.reason = commonSynoErrors[synoerror.code]
	return fmt.Sprintf("Error from Synology API: (%v) %v", synoerror.code, synoerror.reason)
}

func (genericerror *GenericError) Error() string {
	return fmt.Sprintf("Error occured: %v", genericerror.desc)
}
