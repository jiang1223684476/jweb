package app

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
)

// getRequestContext get context from request data
func getRequestContext(data string) Context {
	// context
	context := Context{}

	// header
	context.Header = func(name string, defaultValue string) string {
		compile, _ := regexp.Compile(fmt.Sprintf(`%s:.*`, name))
		value := compile.FindString(data)
		compile, _ = regexp.Compile(fmt.Sprintf(`%s:\s+|\s+$|;\s+boundary.*`, name))
		value = compile.ReplaceAllString(value, "")
		if value != "" {
			return value
		}
		return defaultValue
	}

	// full path
	compile, _ := regexp.Compile(`/.* `)
	fullPath := compile.FindString(data)
	compile, _ = regexp.Compile(`\s+`)
	fullPath = compile.ReplaceAllString(fullPath, "")
	fullPath, _ = url.PathUnescape(fullPath)

	// queries
	var queries []query
	// queries data
	compile, _ = regexp.Compile(`/\?.*`)
	queryData := compile.FindString(fullPath)
	compile, _ = regexp.Compile(`/\?`)
	queryData = compile.ReplaceAllString(queryData, "")
	queryData, _ = url.QueryUnescape(queryData)
	compile, _ = regexp.Compile(`&`)
	// queries split
	querySplit := compile.Split(queryData, -1)
	for _, v := range querySplit {
		compile, _ = regexp.Compile(`=`)
		split := compile.Split(v, -1)

		name := ""
		value := ""
		for k, v := range split {
			if k == 0 {
				name = v
			}
			if k == 1 {
				value = v
			}
		}
		queries = append(queries, query{
			Name:  name,
			Value: value,
		})
	}
	context.Query = func(name string, defaultValue string) string {
		for _, v := range queries {
			if v.Name == name {
				return v.Value
			}
		}
		return defaultValue
	}

	// forms
	var forms []form
	// content type
	contentType := context.Header("Content-Type", "")
	// application/x-www-forms-urlencoded
	if contentType == ContentForm {
		// forms split
		compile, _ = regexp.Compile(`\r\n`)
		formSplit := compile.Split(data, -1)
		// forms data
		formData := formSplit[len(formSplit)-1]
		formData, _ = url.QueryUnescape(formData)
		compile, _ = regexp.Compile(`&`)
		formSplit = compile.Split(formData, -1)
		for _, v := range formSplit {
			compile, _ = regexp.Compile(`=`)
			split := compile.Split(v, -1)

			name := ""
			value := ""
			for k, v := range split {
				if k == 0 {
					name = v
				}
				if k == 1 {
					value = v
				}
			}

			forms = append(forms, form{
				Name:  name,
				Value: value,
			})
		}
	}
	// multipart/forms-data
	if contentType == ContentFormMultipart {
		// forms data
		compile, _ = regexp.Compile(`boundary=[\s\S]*`)
		formData := compile.FindString(data)
		compile, _ = regexp.Compile(`boundary=.*|Content-Length:.*`)
		formData = compile.ReplaceAllString(formData, "")
		compile, _ = regexp.Compile(`---.*`)

		// forms split
		formSplit := compile.Split(formData, -1)
		for _, v := range formSplit {
			if v == "" {
				continue
			}
			// name
			compile, _ = regexp.Compile(`name=".*?"`)
			name := compile.FindString(v)
			compile, _ = regexp.Compile(`name=|"`)
			name = compile.ReplaceAllString(name, "")

			// value
			compile, _ = regexp.Compile(`Content-.*`)
			value := compile.ReplaceAllString(v, "")
			compile, _ = regexp.Compile(`^\s+|\s+$`)
			value = compile.ReplaceAllString(value, "")

			// file name
			compile, _ = regexp.Compile(`filename=".*?"`)
			fileName := compile.FindString(v)
			compile, _ = regexp.Compile(`filename=|"`)
			fileName = compile.ReplaceAllString(fileName, "")

			// file content type
			compile, _ = regexp.Compile(`Content-Type:.*`)
			fileContentType := compile.FindString(v)
			compile, _ = regexp.Compile(`Content-Type:|\s`)
			fileContentType = compile.ReplaceAllString(fileContentType, "")

			forms = append(forms, form{
				Name:        name,
				Value:       value,
				FileName:    fileName,
				ContentType: fileContentType,
			})
		}
	}
	context.Form = func(name string, defaultValue string) string {
		for _, v := range forms {

			if v.Name == name {
				return v.Value
			}
		}
		return defaultValue
	}
	// forms file name
	context.FormFileName = func(name string, defaultValue string) string {
		for _, v := range forms {
			if v.Name == name {
				return v.FileName
			}
		}
		return defaultValue
	}
	// forms file content type
	context.FormFileContentType = func(name string, defaultValue string) string {
		for _, v := range forms {
			if v.Name == name {
				return v.ContentType
			}
		}
		return defaultValue
	}

	// cookie
	context.Cookie = func(name string, defaultValue string) string {
		cookie := context.Header("Cookie", "")

		compile, _ = regexp.Compile(`;`)
		cookieSplit := compile.Split(cookie, -1)
		for _, v := range cookieSplit {
			compile, _ = regexp.Compile(`^\s+|\s+$`)
			v = compile.ReplaceAllString(v, "")

			// cookie name
			compile, _ = regexp.Compile(`=.*`)
			cookieName := compile.ReplaceAllString(v, "")
			// cookie value
			compile, _ = regexp.Compile(`.*?=`)
			cookieValue := compile.ReplaceAllString(v, "")

			if cookieName == name {
				return cookieValue
			}
		}
		return defaultValue
	}

	// session
	context.Session = func(name string, defaultValue string) string {
		return context.Cookie(name, defaultValue)
	}

	// method
	compile, _ = regexp.Compile(
		fmt.Sprintf("^(%s|%s|%s|%s|%s)",
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		),
	)
	method := compile.FindString(data)
	context.Method = method

	// path
	compile, _ = regexp.Compile(`\?.*`)
	requestPath := compile.ReplaceAllString(fullPath, "")
	context.Path = requestPath

	return context
}

// filterRequest filter request
func filterRequest(router router, context *Context) bool {
	// set default response handler
	ResponseHandler(context)

	// exec filters
	if filterBase(router, context) || filterParams(router, context) {
		// exec global handler func
		GlobalHandler(context)
		// exec middleware func
		for _, v := range router.MiddleWares {
			v(context)
		}
		// exec handler func to set context data
		router.Handler(context)

		return true
	}
	if filterStatic(router, context) {
		return true
	}

	// set 404 not found response handler
	NotFoundHandler(context)

	return false
}

// filterBase filter base
func filterBase(router router, context *Context) bool {
	if router.Path == context.Path && router.Method == context.Method {
		return true
	}

	return false
}

// filterParams filter params
func filterParams(router router, context *Context) bool {
	// params
	// /user/name:\w+/id:\d{1}
	compile, _ := regexp.Compile(`/\w+:.*`)
	params := compile.FindString(router.Path)
	if params != "" && router.Method == context.Method {
		// params split by /
		compile, _ = regexp.Compile(`/`)
		paramSplit := compile.Split(params, -1)
		for _, v := range paramSplit {
			// matched :
			matched, _ := regexp.MatchString(`:`, v)
			if matched {
				compile, _ = regexp.Compile(`:`)
				split := compile.Split(v, -1)

				// param struct
				paramStruct := param{
					Name:  "",
					Value: "",
				}
				// set param struct
				for k, v := range split {
					if k == 0 {
						paramStruct.Name = v
					}
					if k == 1 {
						paramStruct.Value = v
					}
				}
				// append to raw params
				context.params = append(context.params, paramStruct)
			}
		}

		// matched path
		// /user/miku/1
		// /user/\w+/\d{1}
		matchedPath := router.Path
		for _, v := range context.params {
			compile, _ = regexp.Compile(fmt.Sprintf("%s|:", v.Name))
			matchedPath = compile.ReplaceAllString(matchedPath, "")
		}

		// regexp matched
		// /user/\w+/\d{1}
		matched, _ := regexp.MatchString(fmt.Sprintf("%s$", matchedPath), context.Path)
		if matched {
			// add ( and ) in matched path to find string sub match
			for _, v := range context.params {
				matchedPath = strings.ReplaceAll(matchedPath, v.Value, fmt.Sprintf("(%s)", v.Value))
			}

			// find string sub match
			compile, _ = regexp.Compile(matchedPath)
			subMatch := compile.FindStringSubmatch(context.Path)
			for k, v := range subMatch {
				// pass first sub match
				if k == 0 {
					continue
				}
				// replace regexp value to path matched value
				context.params[k-1].Value = v
			}
			return true
		}
	}

	return false
}

// filterStatic filter static
func filterStatic(router router, context *Context) bool {
	// check if static request
	compile, _ := regexp.Compile(fmt.Sprintf("^/%s/.*", StaticDirectory))
	staticPath := compile.FindString(context.Path)
	if staticPath != "" && context.Method == http.MethodGet {
		// read file data
		bytes, err := os.ReadFile(fmt.Sprintf(".%s", staticPath))
		if err != nil {
			log.Println(err)
			return false
		}

		// set context data
		context.Data = string(bytes)
		// set content type
		ext := path.Ext(staticPath)
		for k, v := range staticFileTypes {
			if k == ext {
				context.ContentType = v
				break
			}
		}
		return true
	}

	return false
}

// Get request GET
// middleWares before request to exec
func Get(path string, handler func(c *Context), middleWares ...func(c *Context)) {
	addRouters(path, http.MethodGet, handler, middleWares...)
}

// Post request POST
// middleWares before request to exec
func Post(path string, handler func(c *Context), middleWares ...func(c *Context)) {
	addRouters(path, http.MethodPost, handler, middleWares...)
}

// Put request PUT
// middleWares before request to exec
func Put(path string, handler func(c *Context), middleWares ...func(c *Context)) {
	addRouters(path, http.MethodPut, handler, middleWares...)
}

// Delete request DELETE
// middleWares before request to exec
func Delete(path string, handler func(c *Context), middleWares ...func(c *Context)) {
	addRouters(path, http.MethodDelete, handler, middleWares...)
}

// addRouters append routers
func addRouters(path string, method string, handler func(c *Context), middleWare ...func(c *Context)) {
	routers = append(routers, router{
		Path:        path,
		Method:      method,
		Handler:     handler,
		MiddleWares: middleWare,
	})
}

// Params get value form dynamic routing
func Params(context *Context, name string, defaultValue string) string {
	for _, v := range context.params {
		if v.Name == name {
			return v.Value
		}
	}

	return defaultValue
}
