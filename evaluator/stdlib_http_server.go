package evaluator

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/zacanger/cozy/object"
)

var httpServerEnv *object.Environment

type httpRoute struct {
	Pattern *regexp.Regexp
	Handler *object.Function
	Methods []string
}

var routes = make([]httpRoute, 0)

type app struct {
	Routes []httpRoute
}

func newApp() *app {
	app := &app{}
	return app
}

var appInstance = newApp()

func sendWrapper(ctx *httpContext, statusCode int, body string, contentType string) {
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

func registerRoute(env *object.Environment, args ...object.Object) object.Object {
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

func httpContextToCozyReq(c *httpContext) object.Object {
	cReq := make(map[object.HashKey]object.HashPair)
	originalReq := c.Request

	// body
	if originalReq.Body != nil {
		cReqBodyKey := &object.String{Value: "body"}
		buf := new(strings.Builder)
		_, err := io.Copy(buf, originalReq.Body)
		if err != nil {
			return NewError("error in body!, %s", err.Error())
		}
		cReqBodyVal := &object.String{Value: buf.String()}
		cReq[cReqBodyKey.HashKey()] = object.HashPair{Key: cReqBodyKey, Value: cReqBodyVal}
	}

	// content-length
	cReqContentLengthKey := &object.String{Value: "content_length"}
	cReqContentLengthVal := &object.Integer{Value: originalReq.ContentLength}
	cReq[cReqContentLengthKey.HashKey()] = object.HashPair{Key: cReqContentLengthKey, Value: cReqContentLengthVal}

	// method
	cReqMethodKey := &object.String{Value: "method"}
	cReqMethodVal := &object.String{Value: originalReq.Method}
	cReq[cReqMethodKey.HashKey()] = object.HashPair{Key: cReqMethodKey, Value: cReqMethodVal}

	// headers
	cReqHeaders := make(map[object.HashKey]object.HashPair)
	for k, v := range originalReq.Header {
		key := &object.String{Value: k}
		val := &object.String{Value: strings.Join(v, ",")}
		cReqHeaders[key.HashKey()] = object.HashPair{Key: key, Value: val}

	}
	cReqHeadersKey := &object.String{Value: "headers"}
	cReqHeadersVal := &object.Hash{Pairs: cReqHeaders}
	cReq[cReqHeadersKey.HashKey()] = object.HashPair{Key: cReqHeadersKey, Value: cReqHeadersVal}

	cReqContentTypeKey := &object.String{Value: "content_type"}
	cReqContentTypeVal := &object.String{Value: originalReq.Header.Get("Content-Type")}
	cReq[cReqContentTypeKey.HashKey()] = object.HashPair{Key: cReqContentTypeKey, Value: cReqContentTypeVal}

	// url
	cReqURLKey := &object.String{Value: "url"}
	cReqURLVal := &object.String{Value: string(originalReq.URL.String())}
	cReq[cReqURLKey.HashKey()] = object.HashPair{Key: cReqURLKey, Value: cReqURLVal}

	// params
	if c.Params != nil {
		arr := make([]object.Object, 0)
		for _, el := range c.Params {
			arr = append(arr, &object.String{Value: el})
		}
		cReqParamsKey := &object.String{Value: "params"}
		cReqParamsVal := &object.Array{Elements: arr}
		cReq[cReqParamsKey.HashKey()] = object.HashPair{Key: cReqParamsKey, Value: cReqParamsVal}
	}

	// query string
	cReqQuery := make(map[object.HashKey]object.HashPair)
	for k, v := range originalReq.URL.Query() {
		key := &object.String{Value: k}
		val := &object.String{Value: strings.Join(v, ",")}
		cReqQuery[key.HashKey()] = object.HashPair{Key: key, Value: val}
	}
	cReqQueryKey := &object.String{Value: "query"}
	cReqQueryVal := &object.Hash{Pairs: cReqQuery}
	cReq[cReqQueryKey.HashKey()] = object.HashPair{Key: cReqQueryKey, Value: cReqQueryVal}

	return &object.Hash{Pairs: cReq}
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
					applyArgs := make([]object.Object, 0)
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
							emptyPairs := make(map[object.HashKey]object.HashPair)
							headers = &object.Hash{Pairs: emptyPairs}
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

	// TODO: this still isn't quite right
	// we need to disable file directory listings
	// http.Handle(h.Mount, h.Handler).ServeHTTP(w, r)
	// Also it falls through to both 404 and 405 sometimes?
	for _, h := range staticHandlers {
		if strings.HasPrefix(ctx.URL.Path, h.Mount) {
			http.FileServer(http.Dir(h.Path)).ServeHTTP(w, r)
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

type staticHandlerMount struct {
	Mount string
	Path  string
}

var staticHandlers = make([]staticHandlerMount, 0)

// static("./public")
// static("./public", "/some-mount-point")
func staticHandler(env *object.Environment, args ...object.Object) object.Object {
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

func (c *httpContext) send(args ...object.Object) object.Object {
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
	io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
	return &object.Boolean{Value: true}
}

func listen(env *object.Environment, args ...object.Object) object.Object {
	switch a := args[0].(type) {
	case *object.Integer:
		err := http.ListenAndServe(":"+fmt.Sprint(a.Value), appInstance)
		if err != nil {
			return NewError("Could not start server: %s\n", err.Error())
		}
		return &object.Boolean{Value: true}
	default:
		return NewError("http.server.listen expected int port!")
	}
}

func httpServer(env *object.Environment, args ...object.Object) object.Object {
	httpServerEnv = env
	res := make(map[object.HashKey]object.HashPair)

	listenKey := &object.String{Value: "listen"}
	listenVal := &object.Builtin{Fn: listen}
	res[listenKey.HashKey()] = object.HashPair{Key: listenKey, Value: listenVal}

	routeKey := &object.String{Value: "route"}
	routeVal := &object.Builtin{Fn: registerRoute}
	res[routeKey.HashKey()] = object.HashPair{Key: routeKey, Value: routeVal}

	staticKey := &object.String{Value: "static"}
	staticVal := &object.Builtin{Fn: staticHandler}
	res[staticKey.HashKey()] = object.HashPair{Key: staticKey, Value: staticVal}

	return &object.Hash{Pairs: res}
}

func init() {
	RegisterBuiltin("http.create_server",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (httpServer(env, args...))
		})
}
