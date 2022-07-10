package main

import (
	"fmt"
	"net/http"
	"path"
	"sync"
)

type node struct {
	path     string
	children []*node // child nodes, at most 1 :param style node at the end of the array
}

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

type HandlerFunc func(*Context)

type HandlersChain []HandlerFunc

type RouterGroup struct {
	Handlers HandlersChain
	basePath string
	engine   *Engine
	root     bool
}

type Engine struct {
	RouterGroup
	pool  sync.Pool
	trees methodTrees
}

type Param struct {
	Key   string
	Value string
}

type Params []Param

type Context struct {
	engine   *Engine
	params   *Params
	Request  *http.Request
	handlers HandlersChain
	index    int
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

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func (n *node) addRoute(path string, handlers HandlersChain) {
	//fullPath := path

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		//copy(n.children, &node{})
		//n.insertChild(path, fullPath, handlers)
		return
	}

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

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := path.Join(group.basePath, relativePath)
	copy(make(HandlersChain, len(handlers)), handlers)
	copy(handlers, group.Handlers)
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

func (engine *Engine) handleHTTPRequest(c *Context) {
	c.index = -1
	//t := engine.trees
	c.handlers = engine.Handlers
	c.Next()
	c.index = -1
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
	c.Abort()
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

	//r.GET("/myHandler", myHandler)
	//r.POST("/myHandler1", myHandler)

	r.Run(":8888")
}
