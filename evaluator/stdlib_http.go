package evaluator

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"regexp"

	"github.com/zacanger/cozy/object"
)

type httpHandler func(*httpContext)

type httpRoute struct {
	Pattern *regexp.Regexp
	Handler httpHandler
	Methods []string
}

type app struct {
	DefaultRoute httpHandler
	Routes       []httpRoute
}

var appInstances = make(map[int64]app)

func newApp() *app {
	app := &app{
		DefaultRoute: func(ctx *httpContext) {
			ctx.send(&object.Integer{Value: int64(http.StatusNotFound)}, &object.String{Value: "Not found"})
		},
	}

	return app
}

func handleRoute(env *object.Environment, args ...object.Object) object.Object {
	var pattern string
	var methods []string
	var f *object.Function
	var handler httpHandler
	var appInstanceID int64

	switch a := args[0].(type) {
	case *object.Integer:
		appInstanceID = a.Value
	default:
		return NewError("route expected app instance id!")
	}

	switch a := args[1].(type) {
	case *object.String:
		pattern = a.Value
	default:
		return NewError("route expected pattern string!")
	}

	switch a := args[2].(type) {
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

	switch a := args[3].(type) {
	case *object.Function:
		f = a
	default:
		return NewError("route expected callback function!")
	}
	fmt.Println(f)

	appInstance := appInstances[appInstanceID]
	re := regexp.MustCompile(pattern)
	// TODO: get handler out of object.Function f
	route := httpRoute{Pattern: re, Handler: handler, Methods: methods}

	appInstance.Routes = append(appInstance.Routes, route)
	return &object.Boolean{Value: true}
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := &httpContext{Request: r, ResponseWriter: w}

	for _, rt := range a.Routes {
		if matches := rt.Pattern.FindStringSubmatch(ctx.URL.Path); len(matches) > 0 {
			if len(matches) > 1 {
				ctx.Params = matches[1:]
			}

			// TODO: check method against rt.Methods
			rt.Handler(ctx)
			return
		}
	}

	a.DefaultRoute(ctx)
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
	mount := ""
	/*
		var appInstance app

		switch a := args[0].(type) {
		case *object.Integer:
			appInstance = appInstances[a.Value]
		default:
			return NewError("http static expected an app instance id!")
		}
	*/

	switch a := args[1].(type) {
	case *object.String:
		dir = a.Value
	default:
		return NewError("http static expected a string!")
	}

	if len(args) > 2 {
		switch a := args[2].(type) {
		case *object.String:
			mount = a.Value
		}

		if mount == "" {
			mount = "/"
		}
	}

	http.Handle(mount, http.FileServer(http.Dir(dir)))
	return &object.Boolean{Value: true}
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

	app.handleRoute(`^/hello$`, func(ctx *httpContext) {
		ctx.send(http.StatusOK, "Hello world")
	})

	app.handleRoute(`/hello/([\w\._-]+)$`, func(ctx *httpContext) {
		ctx.send(http.StatusOK, fmt.Sprintf("Hello %s", ctx.Params[0]))
	})

	err := http.ListenAndServe(":9000", app)

	if err != nil {
		log.Fatalf("Could not start server: %s\n", err.Error())
	}

}
*/

func listen(env *object.Environment, args ...object.Object) object.Object {
	var appInstance app

	switch a := args[0].(type) {
	case *object.Integer:
		appInstance = appInstances[a.Value]
	default:
		return NewError("http.server.listen expected app instance id!")
	}
	switch a := args[1].(type) {
	case *object.Integer:
		http.ListenAndServe(":"+fmt.Sprint(a.Value), &appInstance)
		return &object.Boolean{Value: true}
	default:
		return NewError("http.server.listen expected int port!")
	}
}

func httpServer(args ...object.Object) object.Object {
	res := make(map[object.HashKey]object.HashPair)

	listenKey := &object.String{Value: "listen"}
	listenVal := &object.Builtin{Fn: listen}
	res[listenKey.HashKey()] = object.HashPair{Key: listenKey, Value: listenVal}

	// route(pattern, callback(context) { respond })
	routeKey := &object.String{Value: "route"}
	routeVal := &object.Builtin{Fn: handleRoute}
	res[routeKey.HashKey()] = object.HashPair{Key: routeKey, Value: routeVal}

	staticKey := &object.String{Value: "static"}
	staticVal := &object.Builtin{Fn: staticHandler}
	res[staticKey.HashKey()] = object.HashPair{Key: staticKey, Value: staticVal}

	appIDKey := &object.String{Value: "id"}
	appIDVal := &object.Integer{Value: rand.Int63()}
	res[appIDKey.HashKey()] = object.HashPair{Key: appIDKey, Value: appIDVal}

	return &object.Hash{Pairs: res}
}

func init() {
	RegisterBuiltin("http.create_server",
		func(env *object.Environment, args ...object.Object) object.Object {
			return (httpServer(args...))
		})
}
