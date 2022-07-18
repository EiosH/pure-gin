package main

import (
	"net/http"
	"sync"
)

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

type HandlersChain []HandlerFunc

type Context struct {
	engine   *Engine
	params   *Params
	Request  *http.Request
	handlers HandlersChain
	index    int
	Params   Params
}
