package synoclient

//"github.com/pkg/errors"

var AuthSynoErrors = map[int]string{
	400: "No such account or incorrect password",
	401: "Account disabled",
	402: "Permission denied",
	403: "2-step verification code required",
	404: "Failed to authenticate 2-step verification code",
}

func (c *Client) Login() (sid string, err error) {
	loginParams := map[string]string{
		"api":     "SYNO.API.Auth",
		"version": "2",
		"method":  "login",
		"account": c.Username,
		"passwd":  c.Password,
		"session": c.Session,
		"format":  "sid",
	}
	resp, err := c.Get("webapi/auth.cgi", loginParams)
	if err != nil {
		return "", HandleApplicationError(resp, err, AuthSynoErrors)
	}

	data := c.GetData(resp)
	sid = data.(map[string]interface{})["sid"].(string)

	// set the sid field to pass with any subsequent request
	c.Sid = sid
	return sid, nil

}

func (c *Client) Logout() error {
	logoutParams := map[string]string{
		"api":     "SYNO.API.Auth",
		"version": "2",
		"method":  "logout",
		"session": c.Session,
	}

	resp, err := c.Get("webapi/auth.cgi", logoutParams)
	if err != nil {
		return HandleApplicationError(resp, err, AuthSynoErrors)
	}

	return nil
}
