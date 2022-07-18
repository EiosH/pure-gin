package main

import (
	"fmt"
	"net/http"
	"path"
)

type node struct {
	path     string
	children []*node
	handlers HandlersChain
	fullPath string
	indices  string
}

type HandlerFunc func(*Context)

type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

type IRoutes interface {
	Use(...HandlerFunc) IRoutes

	GET(string, ...HandlerFunc) IRoutes
	POST(string, ...HandlerFunc) IRoutes
}

func (engine *Engine) allocateContext() *Context {
	v := make(Params, 0)
	return &Context{engine: engine, params: &v}
}

func New() *Engine {
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		trees: make(methodTrees, 0),
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (engine *Engine) Use(middleware ...HandlerFunc) IRoutes {
	engine.RouterGroup.Use(middleware...)
	return engine
}

// POST is a shortcut for router.Handle("POST", path, handle).
func (group *RouterGroup) POST(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodPost, relativePath, handlers)
}

// GET is a shortcut for router.Handle("GET", path, handle).
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodGet, relativePath, handlers)
}

func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		//root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}

	root.addRoute(path, handlers)
}

func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	return path.Join(group.basePath, relativePath)
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	mergedHandlers := make(HandlersChain, len(handlers)+len(group.Handlers))
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group
}

func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group
}

func Default() *Engine {
	engine := New()
	return engine
}

func (engine *Engine) Run(addr string) (err error) {
	defer func() { fmt.Println(err) }()
	err = http.ListenAndServe(addr, engine)
	return
}

func (c *Context) Reset() {
	c.index = -1
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	rPath := c.Request.URL.Path
	httpMethod := c.Request.Method

	t := engine.trees
	for _, node := range t {
		if node.method == httpMethod {
			root := node.root
			v := root.getValue(rPath)
			if v != nil {

				c.handlers = v.handlers
				c.Next()
			}

			break
		}

	}

}
func (c *Context) Next() {
	c.index++

	for c.index < len(c.handlers) {
		c.handlers[c.index](c)
		c.index++
	}

}

func (c *Context) Abort() {
	c.index = len(c.handlers)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.Reset()
	//c.writermem.reset(w)
	c.Request = req
	engine.handleHTTPRequest(c)
	engine.pool.Put(c)
}

func myHandler(c *Context) {
	fmt.Println("myHandler")
	//c.JSON(http.StatusOK, gin.H{
	//	"message": "ok",
	//})
}

func myMiddleware1(c *Context) {
	c.Next()
	fmt.Println("myMiddleware1")
}

func myMiddleware2(c *Context) {
	//c.Abort()
	fmt.Println("myMiddleware2")
}

func myMiddleware3(c *Context) {
	fmt.Println("myMiddleware3")
}

func main() {
	r := Default()
	r.Use(myMiddleware1)
	r.Use(myMiddleware2)
	r.Use(myMiddleware3)

	//api := r.Group("/group", myHandler)
	//api.GET("/a")
	//api.GET("/b")

	r.GET("/a1", myHandler)
	r.GET("/a2", myHandler)

	r.Run(":5678")
}
