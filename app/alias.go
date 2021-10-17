package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

var (
	// MaxRequestSize default max request size is 1MB to receive client data
	MaxRequestSize = 1024 * 1024
	// TemplateDirectory default template directory
	TemplateDirectory = "views"
	// StaticDirectory default static directory
	StaticDirectory = "static"

	// GlobalHandler default global handler
	GlobalHandler = func(c *Context) {
	}

	// ResponseHandler default response handler
	ResponseHandler = func(c *Context) {
		c.Status = http.StatusOK
		c.StatusText = http.StatusText(http.StatusOK)
		c.ContentType = ContentText
	}

	// NotFoundHandler default 404 not found response handler
	NotFoundHandler = func(c *Context) {
		c.Status = http.StatusNotFound
		c.StatusText = http.StatusText(http.StatusNotFound)
		c.ContentType = ContentHTML
		c.Data = http.StatusText(http.StatusNotFound)
	}

	// JsonHandler default json serialization handler
	JsonHandler = func(data interface{}) string {
		marshal, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			return ""
		}

		return string(marshal)
	}

	// HtmlHandler default html rendering handler
	HtmlHandler = func(templateName string, data interface{}) string {
		// the file name extension
		extension := ".html"

		// get template file names
		var templateFileNames []string
		dir, err := ioutil.ReadDir(TemplateDirectory)
		if err != nil {
			log.Println(err)
			return ""
		}
		for _, v := range dir {
			fileName := v.Name()
			// check extension
			ext := path.Ext(fileName)
			if ext == extension {
				templateFileNames = append(templateFileNames, fmt.Sprintf("%s/%s", TemplateDirectory, fileName))
			}
		}

		// parse template file name
		t := template.New(fmt.Sprintf("%s%s", templateName, extension))
		// custom func
		t.Funcs(template.FuncMap{})
		// parse files
		_, err = t.ParseFiles(templateFileNames...)
		if err != nil {
			log.Println(err)
			return ""
		}
		// get template parse data
		buffer := new(bytes.Buffer)
		err = t.Execute(buffer, data)
		if err != nil {
			log.Println(err)
			return ""
		}

		return buffer.String()
	}

	// routers save router array
	routers []router

	// staticFileTypes static file types
	staticFileTypes = map[string]string{
		".avif": "image/avif",
		".css":  "text/css; charset=utf-8",
		".gif":  "image/gif",
		".htm":  "text/html; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".jpeg": "image/jpeg",
		".jpg":  "image/jpeg",
		".js":   "text/javascript; charset=utf-8",
		".json": "application/json",
		".mjs":  "text/javascript; charset=utf-8",
		".pdf":  "application/pdf",
		".png":  "image/png",
		".svg":  "image/svg+xml",
		".wasm": "application/wasm",
		".webp": "image/webp",
		".xml":  "text/xml; charset=utf-8",
	}
)

type (
	// Context to save request amd response data
	Context struct {
		Header              func(name string, defaultValue string) string
		Query               func(name string, defaultValue string) string
		Form                func(name string, defaultValue string) string
		FormFileName        func(name string, defaultValue string) string
		FormFileContentType func(name string, defaultValue string) string
		Cookie              func(name string, defaultValue string) string
		Session             func(name string, defaultValue string) string

		Method      string
		Path        string
		Status      int
		StatusText  string
		ContentType string
		Data        string

		headers []Header
		params  []param
	}

	// router to handler context
	router struct {
		Path        string
		Method      string
		Handler     func(c *Context)
		MiddleWares []func(c *Context)
	}

	// Cookie save cookie data from request or response form server
	Cookie struct {
		Name     string
		Value    string
		MaxAge   int
		Path     string
		Domain   string
		HttpOnly bool
		Secure   bool
		SameSite string
	}

	// Header save header data from request
	Header struct {
		Name     string
		Value    string
		toString string
	}

	// query save query data from request
	query struct {
		Name  string
		Value string
	}

	// form save form data from request
	form struct {
		Name        string
		Value       string
		FileName    string
		ContentType string
	}

	// param save param data from request path
	param struct {
		Name  string
		Value string
	}

	// Map alias for map[string]interface{}
	Map map[string]interface{}

	DBConfig struct {
		User     string
		Password string
		Address  string
		Database string
	}
)

const (
	ContentBinary        = "application/octet-stream"
	ContentWebassembly   = "application/wasm"
	ContentHTML          = "text/html"
	ContentJSON          = "application/json"
	ContentJSONProblem   = "application/problem+json"
	ContentXMLProblem    = "application/problem+xml"
	ContentJavascript    = "text/javascript"
	ContentCssShelt      = "text/javascript"
	ContentText          = "text/plain"
	ContentXML           = "text/xml"
	ContentXMLUnreadable = "application/xml"
	ContentMarkdown      = "text/markdown"
	ContentYAML          = "application/x-yaml"
	ContentYAMLText      = "text/yaml"
	ContentProtobuf      = "application/x-protobuf"
	ContentMsgPack       = "application/msgpack"
	ContentMsgPack2      = "application/x-msgpack"
	ContentForm          = "application/x-www-form-urlencoded"
	ContentFormMultipart = "multipart/form-data"
	ContentGRPC          = "application/grpc"
)
