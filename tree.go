package main

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

func (n *node) addRoute(path string, handlers HandlersChain) {
	fullPath := path

	// 当前不存在 直接插入
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, fullPath, handlers)
		return
	}

	parentFullPathIndex := 0

walk:
	for {

		i := longestCommonPrefix(path, n.path)

		// 公共前缀小于目前已存在的 path , 取公共前缀分裂
		// a1  a2 =>  a - 12
		if i < len(n.path) {
			child1 := node{
				path:     n.path[i:],
				indices:  "",
				children: []*node{},
				handlers: n.handlers,
			}

			child2 := node{
				path:     path[i:],
				indices:  "",
				children: []*node{},
				handlers: handlers,
			}

			n.children = []*node{&child1, &child2}
			n.indices = child1.path + child2.path
			n.path = path[:i]
			n.handlers = nil
			n.fullPath = fullPath[:parentFullPathIndex+i]
		}

		// 公共前缀不小于目前已存在的 path , 取已存在的 path 分裂
		// a1  a12 =>  a1 - 2
		// todo
		if i+10000 < len(path) {
			//	path = path[i:]
			//	c := path[0]
			//
			//	n.insertChild(path, fullPath, handlers)
			continue walk
		}

		return
	}
}

func (n *node) insertChild(path string, fullPath string, handlers HandlersChain) {
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}

func (n *node) getValue(path string) *node {
	prefix := n.path
	resNode := n

walk:
	for {
		if prefix == path {
			return resNode
		} else {
			for _, node := range n.children {
				prefix += node.path
				resNode = node
				if longestCommonPrefix(prefix, path) >= len(path) {
					continue walk
				}
			}

			return nil

		}
	}

}
