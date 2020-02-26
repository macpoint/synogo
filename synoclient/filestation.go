package synoclient

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
)

// "fmt"

type FsSpecificError struct {
	code   int
	reason string
}

var FsSynoErrors = map[int]string{
	1002: "An error occurred at the destination",
	1200: "Failed to rename file",
	// more to come
}

var FsSpecifiErrors = map[int]string{
	400: "Invalid parameter of file operation",
	401: "Unknown error of file operation",
	402: "System is too busy",
	403: "Invalid user does this file operation",
	404: "Invalid group does this file operation",
	405: "Invalid user and group does this file operation",
	406: "Can’t get user/group information from the account server",
	407: "Operation not permitted",
	408: "No such file or directory",
	409: "Non-supported file system",
	410: "Failed to connect internet-based file system",
	411: "Read-only file system",
	412: "Filename too long in the non-encrypted file system",
	413: "Filename too long in the encrypted file system",
	414: "File already exists",
	415: "Disk quota exceeded",
	416: "No space left on device",
	417: "Input/output error",
	418: "Illegal name or path",
	419: "Illegal file name",
	420: "Illegal file name on FAT file system",
	421: "Device or resource busy",
	599: "No such task of the file operation",
}

func specifyError(response string, err error) error {
	var responseData interface{}
	json.Unmarshal([]byte(response), &responseData)
	errorBlock := responseData.(map[string]interface{})["error"]
	// error -> errors{[0]} -> code
	nestedErrorCode := int(errorBlock.(map[string]interface{})["errors"].([]interface{})[0].(map[string]interface{})["code"].(float64))
	return errors.Wrap(err, FsSpecifiErrors[nestedErrorCode])
}

func (fserror *FsSpecificError) Error() string {
	return fmt.Sprintf("File station error: (%v) %v", fserror.code, fserror.reason)
}

func (c *Client) RenameFile(path string, name string) (string, error) {
	params := map[string]string{
		"api":     "SYNO.FileStation.Rename",
		"version": "1",
		"method":  "rename",
		"path":    path,
		"name":    name,
	}

	resp, err := c.Get("webapi/entry.cgi", params)
	if err != nil {
		return "", specifyError(resp, HandleApplicationError(resp, err, FsSynoErrors))
	}

	return string(c.GetData(resp).(map[string]interface{})["files"].([]interface{})[0].(map[string]interface{})["path"].(string)), nil
}

func (c *Client) MoveFile(sourceFile string, destinationDir string) error {
	params := map[string]string{
		"api":              "SYNO.FileStation.CopyMove",
		"version":          "1",
		"method":           "start",
		"path":             sourceFile,
		"dest_folder_path": destinationDir,
		"remove_src":       "true",
	}

	resp, err := c.Get("webapi/entry.cgi", params)
	if err != nil {
		fmt.Println(resp)
		return specifyError(resp, HandleApplicationError(resp, err, FsSynoErrors))
	}

	return nil

}
