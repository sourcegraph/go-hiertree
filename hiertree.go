// Package hiertree arranges a list of flat paths (and associated objects) into a hierarchical tree.
package hiertree

import (
	"errors"
	"sort"
	"strings"
)

// Elem represents an object with a path.
type Elem interface {
	// HierPath is the object's path in the tree, with path components separated by slashes
	// (e.g., "a/b/c").
	HierPath() string
}

// Entry represents an entry in the resulting hierarchical tree. Elem is nil if the entry is a
// stub (i.e., no element exists at that path, but it does contain elements).
type Entry struct {
	// Parent is the full path of this entry's parent
	Parent string

	// Name is the name of this entry (without the parent path)
	Name string

	// Elem is the element that exists at this position in the tree, or nil if the entry is a stub
	Elem Elem
}

// List arranges elems into a flat list based on their hierarchical paths.
func List(elems []Elem) (entries []Entry, err error) {
	var nodes []Node
	nodes, err = Tree(elems)
	if err == nil {
		entries, err = list(nodes, "")
	}
	return
}

func list(nodes []Node, parent string) (entries []Entry, err error) {
	var err2 error
	for _, n := range nodes {
		entries = append(entries, Entry{
			Parent: parent,
			Name:   n.Name,
			Elem:   n.Elem,
		})
		var prefix string
		if parent == "" {
			prefix = n.Name
		} else {
			prefix = parent + "/" + n.Name
		}
		var children []Entry
		children, err2 = list(n.Children, prefix)
		if err2 != nil && err == nil {
			err = err2
		}
		entries = append(entries, children...)
	}
	return
}

// Node represents a node in the resulting tree. Elem is nil if the entry is a stub (i.e., no
// element exists at this path, but it does contain elements).
type Node struct {
	// Name is the name of this node (without the parent path)
	Name string

	// Elem is the element that exists at this position in the tree, or nil if the entry is a stub
	Elem Elem

	// Children is the list of child nodes under this node
	Children []Node
}

// Tree arranges elems into a tree based on their hierarchical paths.
func Tree(elems []Elem) (nodes []Node, err error) {
	nodes, _, err = tree(elems, "")
	return
}

func tree(elems []Elem, prefix string) (roots []Node, size int, err error) {
	var err2 error
	es := elemlist(elems)
	if prefix == "" { // only sort on first call
		sort.Sort(es)
	}
	var cur *Node
	var saveCur = func() {
		if cur != nil {
			size++
			roots = append(roots, *cur)
		}
		cur = nil
	}
	defer saveCur()
	for i := 0; i < len(es); i++ {
		e := es[i]
		path := e.HierPath()
		if !strings.HasPrefix(path, prefix) {
			return
		}
		relpath := path[len(prefix):]
		root, rest := split(relpath)
		if root == "" && err == nil {
			err = errors.New("invalid node path: " + path)
		}
		if cur != nil && cur.Name == relpath && err == nil {
			err = errors.New("duplicate node path: " + path)
		}
		if cur == nil || cur.Name != root {
			saveCur()
			cur = &Node{Name: root}
		}
		if rest == "" {
			cur.Elem = e
		}
		var n int
		cur.Children, n, err2 = tree(elems[i:], prefix+root+"/")
		if err2 != nil && err == nil {
			err = err2
		}
		size += n
		if n > 0 {
			i += n - 1
		}
	}
	return
}

type elemlist []Elem

func (vs elemlist) Len() int           { return len(vs) }
func (vs elemlist) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs elemlist) Less(i, j int) bool { return vs[i].HierPath() < vs[j].HierPath() }

// split splits path immediately following the first slash. The returned values have the property
// that path = root+"/"+rest.
func split(path string) (root, rest string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}
