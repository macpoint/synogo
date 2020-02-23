package synoclient

import "fmt"

// SynoError ...
type SynoError struct {
	code   int
	reason string
}

// GenericError ...
type GenericError struct {
	desc string
}

var commonErrors = map[int]string{
	100: "Unknown error",
	101: "Invalid parameter",
	102: "The requested API does not exist",
	103: "The requested method does not exist",
	104: "The requested version does not support the functionality",
	105: "The logged in session does not have permission",
	106: "Session timeout",
	107: "Session interrupted by duplicate login",
	// auth specific errors
	400: "No such account or incorrect password",
	401: "Account disabled",
	402: "Permission denied",
	403: "2-step verification code required",
	404: "Failed to authenticate 2-step verification code",
}

func (synoerror *SynoError) Error() string {
	message := commonErrors[synoerror.code]
	return fmt.Sprintf("Error from Synology API: (%v) %v", synoerror.code, message)
}

func (genericerror *GenericError) Error() string {
	return fmt.Sprintf("Error occured: %v", genericerror.desc)
}
