package app

import (
	"fmt"
	"net"
	"net/http"
)

// writeData writes data to the connection client
func writeData(conn net.Conn, c Context) {
	// header data
	data := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: %s\r\n",
		c.Status,
		c.StatusText,
		c.ContentType,
	)
	// set header
	for _, v := range c.headers {
		data += fmt.Sprintf("%s\r\n", v.toString)
	}
	// end header data
	data += "\r\n"

	// body data
	data += fmt.Sprintf("%s", c.Data)

	// write data
	_, _ = conn.Write([]byte(data))
	// close connection
	_ = conn.Close()
}

// SetHeader set response header
func SetHeader(context *Context, header Header) {
	// to string
	header.toString = fmt.Sprintf("%s: %s", header.Name, header.Value)
	// append to headers
	context.headers = append(context.headers, header)
}

// SetCookie set response cookie
func SetCookie(context *Context, cookie Cookie) {
	// value
	value := fmt.Sprintf("%s=%s;", cookie.Name, cookie.Value)

	// max age
	if cookie.MaxAge >= 0 {
		value += fmt.Sprintf(" Max-Age=%d;", cookie.MaxAge)
	}
	// path
	if cookie.Path != "" {
		value += fmt.Sprintf(" Path=%s;", cookie.Path)
	}
	// domain
	if cookie.Domain != "" {
		value += fmt.Sprintf(" Domain=%s;", cookie.Domain)
	}
	// http only
	if cookie.HttpOnly {
		value += " HttpOnly;"
	}
	// secure
	if cookie.Secure {
		value += " secure;"

		// same site
		if cookie.SameSite == "" {
			cookie.SameSite = "Lax"
		}
		value += fmt.Sprintf(" SameSite=%s;", cookie.SameSite)
	}

	// Set-Cookie
	SetHeader(context, Header{
		Name:  "Set-Cookie",
		Value: value,
	})
}

// RemoveCookie set remove cookie header
func RemoveCookie(context *Context, name string) {
	SetCookie(context, Cookie{Name: name})
}

// SetSession set response session
func SetSession(context *Context, name string, value string) {
	SetCookie(context, Cookie{Name: name, Value: value, MaxAge: -1})
}

// RemoveSession set remove session header
func RemoveSession(context *Context, name string) {
	SetCookie(context, Cookie{Name: name})
}

// Redirect set redirect header
func Redirect(context *Context, path string) {
	context.Status = http.StatusTemporaryRedirect
	context.StatusText = http.StatusText(http.StatusTemporaryRedirect)

	// Location
	SetHeader(context, Header{
		Name:  "Location",
		Value: path,
	})
}

// Json serialization data to json
func Json(context *Context, data interface{}) string {
	context.ContentType = ContentJSON

	// default json serialization handler
	return JsonHandler(data)
}

// Html rendering data to html
func Html(context *Context, templateName string, data interface{}) string {
	context.ContentType = ContentHTML

	// default html rendering handler
	return HtmlHandler(templateName, data)
}
