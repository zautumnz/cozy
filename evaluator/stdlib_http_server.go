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

	cReqContentLengthKey := &object.String{Value: "content_length"}
	cReqContentLengthVal := &object.Integer{Value: originalReq.ContentLength}
	cReq[cReqContentLengthKey.HashKey()] = object.HashPair{Key: cReqContentLengthKey, Value: cReqContentLengthVal}

	cReqMethodKey := &object.String{Value: "method"}
	cReqMethodVal := &object.String{Value: originalReq.Method}
	cReq[cReqMethodKey.HashKey()] = object.HashPair{Key: cReqMethodKey, Value: cReqMethodVal}

	cReqHeaders := make(map[object.HashKey]object.HashPair)
	for k, v := range originalReq.Header {
		key := &object.String{Value: k}
		val := &object.String{Value: strings.Join(v, ",")}
		cReqHeaders[key.HashKey()] = object.HashPair{Key: key, Value: val}

	}
	cReqHeadersKey := &object.String{Value: "headers"}
	cReqHeadersVal := &object.Hash{Pairs: cReqHeaders}
	cReq[cReqHeadersKey.HashKey()] = object.HashPair{Key: cReqHeadersKey, Value: cReqHeadersVal}

	if c.Params != nil {
		arr := make([]object.Object, 0)
		for _, el := range c.Params {
			arr = append(arr, &object.String{Value: el})
		}
		cReqParamsKey := &object.String{Value: "params"}
		cReqParamsVal := &object.Array{Elements: arr}
		cReq[cReqParamsKey.HashKey()] = object.HashPair{Key: cReqParamsKey, Value: cReqParamsVal}
	}

	// TODO: same as above for each anything else relevant on originalReq

	return &object.Hash{Pairs: cReq}
}

// TODO: make this `a` our `appInstance`
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
					switch res.(type) {
					case *object.Hash:
						a := res.(*object.Hash)
						bodyStr := &object.String{Value: "body"}
						contentTypeStr := &object.String{Value: "content_type"}
						statusCodeStr := &object.String{Value: "status_code"}
						body := a.Pairs[bodyStr.HashKey()].Value
						contentType := a.Pairs[contentTypeStr.HashKey()].Value
						statusCode := a.Pairs[statusCodeStr.HashKey()].Value

						sc := 200
						bd := ""
						ct := "text/plain"

						// TODO: add a headers object
						switch s := statusCode.(type) {
						case *object.Integer:
							sc = int(s.Value)
						default:
							break
						}

						switch b := body.(type) {
						case *object.String:
							bd = b.Value
						default:
							break
						}

						switch c := contentType.(type) {
						case *object.String:
							ct = c.Value
						default:
							break
						}

						sendWrapper(ctx, sc, bd, ct)
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

	notFound(ctx)
}

type httpContext struct {
	http.ResponseWriter
	*http.Request
	Params []string
}

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

	// TODO: return this so it can work with request and response
	http.Handle(mount, http.FileServer(http.Dir(dir)))
	return NULL
}

// this method can be used in the cozy stdlib to build
// all other reponse methods (text, json, etc.)
// example: json = (hash) -> ctx.send(200, json.serialize(hash))
// TODO: need to figure out how to pass the context instance
func (c *httpContext) send(args ...object.Object) object.Object {
	code := 200
	body := ""
	contentType := "text/plain"
	switch a := args[0].(type) {
	case *object.Integer:
		code = int(a.Value)
	default:
		return NewError("Incorrect argument provided to route handler")
	}
	switch a := args[1].(type) {
	case *object.String:
		body = a.Value
	default:
		return NewError("Incorrect argument provided to route handler")
	}
	switch a := args[2].(type) {
	case *object.String:
		contentType = a.Value
	default:
		return NewError("Incorrect argument provided to route handler")

	}

	c.ResponseWriter.Header().Set("Content-Type", contentType)
	c.WriteHeader(code)
	io.WriteString(c.ResponseWriter, fmt.Sprintf("%s\n", body))
	return &object.Boolean{Value: true}
}

/*
func main() {
	app := newApp()

	app.registerRoute(`^/hello$`, func(ctx *httpContext) {
		ctx.send(http.StatusOK, "Hello world")
	})

	app.registerRoute(`/hello/([\w\._-]+)$`, func(ctx *httpContext) {
		ctx.send(http.StatusOK, fmt.Sprintf("Hello %s", ctx.Params[0]))
	})

	err := http.ListenAndServe(":9000", app)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
*/

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

	// route(pattern, callback(context) { respond })
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
