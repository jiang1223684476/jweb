# Jweb
* Http web frame with Go Programming Language

## Install
```

```

## How to use?
```go
package main

import (
	"JWeb/app"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path"
)

func main() {
	// replace default max request size
	app.MaxRequestSize = 1024 * 1024
	// replace default template directory
	app.TemplateDirectory = "views"
	// replace default static directory
	app.StaticDirectory = "static"

	// replace default global handler
	app.GlobalHandler = func(c *app.Context) {
	}
	// replace default response handler
	app.ResponseHandler = func(c *app.Context) {
		c.Status = http.StatusOK
		c.StatusText = http.StatusText(http.StatusOK)
		c.ContentType = app.ContentText
	}
	// replace default not found handler
	app.NotFoundHandler = func(c *app.Context) {
		c.Status = http.StatusNotFound
		c.StatusText = http.StatusText(http.StatusNotFound)
		c.ContentType = app.ContentHTML
		c.Data = http.StatusText(http.StatusNotFound)
	}
	// replace default json serialization handler
	app.JsonHandler = func(data interface{}) string {
		marshal, err := json.Marshal(data)
		if err != nil {
			log.Println(err)
			return ""
		}

		return string(marshal)
	}
	// replace default html rendering handler
	app.HtmlHandler = func(templateName string, data interface{}) string {
		// the file name extension
		extension := ".html"

		// get template file names
		var templateFileNames []string
		dir, err := ioutil.ReadDir(app.TemplateDirectory)
		if err != nil {
			log.Println(err)
			return ""
		}
		for _, v := range dir {
			fileName := v.Name()
			// check extension
			ext := path.Ext(fileName)
			if ext == extension {
				templateFileNames = append(templateFileNames, fmt.Sprintf("%s/%s", app.TemplateDirectory, fileName))
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

	// request GET
	app.Get("/", func(c *app.Context) {
		// app.Html rendering data to html
		c.Data = app.Html(c, "index", app.Map{
			// Path request path
			"path": c.Path,
			// Method request path
			"method": c.Method,
			// Query request query value
			// http://127.0.0.1:8080/?id=1
			"id": c.Query("id", "0"),
		})
	})
	// Dynamic routing using regexp
	// for example /user/id:\d+
	// http://127.0.0.1:8080/user/1
	app.Get(`/user/id:\d+`, func(c *app.Context) {
		// Cookie get cookie
		cookieName := c.Cookie("cookieName", "")
		if cookieName == "" {
			// SetCookie set response cookie
			app.SetCookie(c, app.Cookie{Name: "cookieName", Value: "miku", MaxAge: 60 * 60 * 24})
		} else {
			log.Println(cookieName)
		}
		// RemoveCookie remove cookie by name
		app.RemoveCookie(c, "cookieName")

		// Session get session
		sessionName := c.Session("sessionName", "")
		if sessionName == "" {
			// SetSession set response session
			app.SetSession(c, "sessionName", "haku")
		} else {
			log.Println(sessionName)
		}
		// RemoveSession remove session by name
		app.RemoveSession(c, "sessionName")

		// SetHeader set response header
		app.SetHeader(c, app.Header{
			Name:  "Server",
			Value: "Test",
		})

		// Redirect set redirect header
		//app.Redirect(c, "/")

		// set specify response Content-Type
		//c.ContentType = app.ContentJSON

		// app.Json serialization data to json
		c.Data = app.Json(c, app.Map{
			"path":   c.Path,
			"method": c.Method,
			// Params get value form dynamic routing
			"id": app.Params(c, "id", "0"),
			// get request specify header value
			"User-Agent": c.Header("User-Agent", ""),
		})
	})
	// request POST
	app.Post("/", func(c *app.Context) {
		c.Data = app.Json(c, app.Map{
			"path":   c.Path,
			"method": c.Method,
			"id":     c.Form("id", "0"),
		})
	})
	// request PUT
	app.Put("/", func(c *app.Context) {
		c.Data = app.Json(c, app.Map{
			"path":   c.Path,
			"method": c.Method,
			"id":     c.Form("id", "0"),
		})
	})
	// request DELETE
	app.Delete("/", func(c *app.Context) {
		c.Data = app.Json(c, app.Map{
			"path":   c.Path,
			"method": c.Method,
			"id":     c.Form("id", "0"),
		})
	})

	address := "127.0.0.1:8080"
	// run applications in specify address
	app.Run(address)
}
```

## Have fun!