package synoclient

import (
// "fmt"
)

// TBD

func (c *Client) RenameFile(path string, name string) error {
	params := map[string]string{
		"api":     "SYNO.FileStation.Rename",
		"version": "1",
		"method":  "rename",
		"path":    path,
		"name":    name,
	}

	_, err := c.Get("webapi/entry.cgi", params)
	if err != nil {
		return err
	}

	return nil
}
