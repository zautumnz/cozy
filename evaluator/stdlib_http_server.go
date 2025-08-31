package evaluator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/zautumnz/cozy/object"
)

var httpServerEnv *ENV

type httpRoute struct {
	Pattern *regexp.Regexp
	Handler *object.Function
	Methods []string
}

var routes []httpRoute
var appInstance *app

type app struct {
	Routes []httpRoute
}

func sendWrapper(
	ctx *httpContext,
	statusCode int,
	body string,
	contentType string,
) {
	ctx.send(
		&object.Integer{Value: int64(statusCode)},
		&object.String{Value: body},
		&object.String{Value: contentType},
	)
}

func notFound(ctx *httpContext) {
	sendWrapper(ctx, http.StatusNotFound, "Not found", "text/plain")
}

func methodNotAllowed(ctx *httpContext) {
	sendWrapper(ctx, http.StatusMethodNotAllowed, "Not allowed", "text/plain")
}

func registerRoute(env *ENV, args ...OBJ) OBJ {
	var pattern string
	var methods []string
	var handler *object.Function

	switch a := args[0].(type) {
	case *object.String:
		pattern = a.Value
	default:
		return NewError("route expected pattern string!")
	}

	switch a := args[1].(type) {
	case *object.Array:
		for _, e := range a.Elements {
			switch x := e.(type) {
			case *object.String:
				methods = append(methods, x.Value)
			default:
				return NewError("route expected methods string array!")
			}
		}
	default:
		return NewError("route expected methods string array!")
	}

	switch f := args[2].(type) {
	case *object.Function:
		handler = f
	default:
		return NewError("route expected callback function!")
	}

	re := regexp.MustCompile(pattern)
	route := httpRoute{Pattern: re, Handler: handler, Methods: methods}

	routes = append(routes, route)
	return NULL
}

func httpReqFormToCozyForm(c *httpContext) OBJ {
	formValues := make(StringObjectMap)
	for k, v := range c.Request.PostForm {
		formValues[k] = &object.String{Value: strings.Join(v, ",")}
	}
	return NewHash(formValues)
}

func httpContextToCozyReq(c *httpContext) OBJ {
	cReq := make(StringObjectMap)
	originalReq := c.Request

	originalContentType := originalReq.Header.Get("Content-Type")
	if strings.HasPrefix(originalContentType, "multipart/form-data") {
		// 50 mb max memory; anything over this is written to a tmp file
		// This could be configurable in the future
		originalReq.ParseMultipartForm(50 << 20)
		filesMap := make(StringObjectMap)

		for inputName, files := range originalReq.MultipartForm.File {
			filesArr := make([]OBJ, 0)

			for _, fileHeader := range files {
				file, err := fileHeader.Open()
				if err != nil {
					return NewError("error opening uploaded file header")
				}
				defer file.Close()

				buff := make([]byte, 512)
				_, err = file.Read(buff)
				if err != nil {
					return NewError("error reading uploaded file")
				}

				_, err = file.Seek(0, io.SeekStart)
				if err != nil {
					return NewError("error in reading uploaded file")
				}

				t := os.TempDir()
				dir := t + "/cozy/http/uploads"
				mode, _ := strconv.ParseInt("755", 8, 64)
				err = os.MkdirAll(dir, os.FileMode(mode))
				if err != nil {
					return NewError("error ensuring temp dir")
				}

				p := fmt.Sprintf(
					"%s/%d%s",
					dir,
					time.Now().UnixNano(),
					filepath.Ext(fileHeader.Filename),
				)
				f, err := os.Create(p)
				if err != nil {
					return NewError("error saving uploaded file")
				}
				defer f.Close()

				_, err = io.Copy(f, file)
				if err != nil {
					return NewError("error saving uploaded file")
				}

				fMap := NewHash(StringObjectMap{
					"name": &object.String{Value: fileHeader.Filename},
					"file": &object.File{Filename: p},
				})
				filesArr = append(filesArr, fMap)
			}

			filesMap[inputName] = &object.Array{Elements: filesArr}
		}

		cReq["files"] = NewHash(filesMap)
		cReq["form"] = httpReqFormToCozyForm(c)
	} else if originalContentType == "application/x-www-form-urlencoded" {
		originalReq.ParseForm()
		cReq["form"] = httpReqFormToCozyForm(c)
	} else if originalReq.Body != nil {
		// we don't grab a body if there's a form, because otherwise we'd end up
		// with form values, including form-data/uploads, on the body hash.
		buf := new(strings.Builder)
		_, err := io.Copy(buf, originalReq.Body)
		if err != nil {
			return NewError("error in body!, %s", err.Error())
		}
		cReq["body"] = &object.String{Value: buf.String()}
	}

	cReq["content_length"] = &object.Integer{Value: originalReq.ContentLength}
	cReq["content_type"] = &object.String{Value: originalContentType}
	cReq["method"] = &object.String{Value: originalReq.Method}
	cReq["url"] = &object.String{Value: string(originalReq.URL.String())}

	// headers
	cReqHeaders := make(StringObjectMap)
	for k, v := range originalReq.Header {
		cReqHeaders[k] = &object.String{Value: strings.Join(v, ",")}
	}
	cReq["headers"] = NewHash(cReqHeaders)

	// params
	if c.Params != nil {
		arr := make([]OBJ, 0)
		for _, el := range c.Params {
			arr = append(arr, &object.String{Value: el})
		}
		cReq["params"] = &object.Array{Elements: arr}
	}

	// query string
	cReqQuery := make(StringObjectMap)
	for k, v := range originalReq.URL.Query() {
		cReqQuery[k] = &object.String{Value: strings.Join(v, ",")}
	}
	cReq["query"] = NewHash(cReqQuery)

	return NewHash(cReq)
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &httpContext{Request: r, ResponseWriter: w}

	for _, rt := range routes {
		if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}

			for _, m := range rt.Methods {
				if m == r.Method {
					applyArgs := make([]OBJ, 0)
					applyArgs = append(applyArgs, httpContextToCozyReq(ctx))
					res := ApplyFunction(httpServerEnv, rt.Handler, applyArgs)
					switch a := res.(type) {
					case *object.Hash:
						bodyStr := &object.String{Value: "body"}
						contentTypeStr := &object.String{Value: "content_type"}
						statusCodeStr := &object.String{Value: "status_code"}
						body := a.Pairs[bodyStr.HashKey()].Value
						contentType := a.Pairs[contentTypeStr.HashKey()].Value
						statusCode := a.Pairs[statusCodeStr.HashKey()].Value
						headersStr := &object.String{Value: "headers"}
						headers := a.Pairs[headersStr.HashKey()].Value

						if statusCode == nil {
							statusCode = &object.Integer{Value: 200}
						}
						if body == nil {
							body = &object.String{Value: ""}
						}
						if contentType == nil {
							contentType = &object.String{Value: "text/plain"}
						}
						if headers == nil {
							headers = NewHash(StringObjectMap{})
						}
						ctx.send(statusCode, body, contentType, headers)
					default:
						fmt.Println(res.Type(), "\n\noh no", res)
						return
					}
					return
				}
			}

			methodNotAllowed(ctx)
		}
	}

	for _, h := range staticHandlers {
		if strings.HasPrefix(ctx.URL.Path, h.Mount) {
			http.FileServer(neuteredFileSystem{http.Dir(h.Path)}).ServeHTTP(w, r)
			return
		}
	}

	notFound(ctx)
}

type httpContext struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

type neuteredFileSystem struct {
	fs http.FileSystem
}

func (nfs neuteredFileSystem) Open(path string) (http.File, error) {
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}

	s, _ := f.Stat()
	if s.IsDir() {
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return f, nil
}

type staticHandlerMount struct {
	Mount string
	Path  string
}

var staticHandlers []staticHandlerMount

// static("./public")
// static("./public", "/some-mount-point")
func staticHandler(env *ENV, args ...OBJ) OBJ {
	dir := ""
	mount := "/"

	switch a := args[0].(type) {
	case *object.String:
		dir = a.Value
	default:
		return NewError("http static expected a string!")
	}

	if len(args) > 1 {
		switch a := args[1].(type) {
		case *object.String:
			mount = a.Value
		}

		if mount == "" {
			mount = "/"
		}
	}

	staticHandlers = append(staticHandlers, staticHandlerMount{
		Mount: mount,
		Path:  dir,
	})

	return NULL
}

func (c *httpContext) send(args ...OBJ) OBJ {
	code := 200
	body := ""
	contentType := "text/plain"
	extraHeaders := make(map[string]string)
	switch a := args[0].(type) {
	case *object.Integer:
		code = int(a.Value)
	default:
		return NewError("Incorrect argument provided to route handler 1")
	}
	switch a := args[1].(type) {
	case *object.String:
		body = a.Value
	default:
		return NewError("Incorrect argument provided to route handler 2")
	}
	switch a := args[2].(type) {
	case *object.String:
		contentType = a.Value
	default:
		return NewError("Incorrect argument provided to route handler 3")

	}
	if len(args) > 3 {
		switch a := args[3].(type) {
		case *object.Hash:
			for _, pair := range a.Pairs {
				extraHeaders[pair.Key.Inspect()] = pair.Value.Inspect()
			}
		default:
			return NewError("Incorrect argument provided to route handler 4")
		}
	}

	c.ResponseWriter.Header().Set("Content-Type", contentType)
	for k, v := range extraHeaders {
		c.ResponseWriter.Header().Set(k, v)
	}
	c.WriteHeader(code)
	if body != "" {
		io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
	}
	return NULL
}

func listen(env *ENV, args ...OBJ) OBJ {
	switch a := args[0].(type) {
	case *object.Integer:
		err := http.ListenAndServe(":"+fmt.Sprint(a.Value), appInstance)
		if err != nil {
			return NewError("Could not start server: %s\n", err.Error())
		}
		return NULL
	default:
		return NewError("http.server.listen expected int port!")
	}
}

func httpServer(env *ENV, args ...OBJ) OBJ {
	httpServerEnv = env

	return NewHash(StringObjectMap{
		"listen": &object.Builtin{Fn: listen},
		"route":  &object.Builtin{Fn: registerRoute},
		"static": &object.Builtin{Fn: staticHandler},
	})
}

func init() {
	routes = make([]httpRoute, 0)
	appInstance = &app{}
	staticHandlers = make([]staticHandlerMount, 0)

	RegisterBuiltin("http.create_server",
		func(env *ENV, args ...OBJ) OBJ {
			return httpServer(env, args...)
		})
}
