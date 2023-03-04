package xml2json

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"strconv"
	"strings"
)

func Convert(b []byte) ([]byte, error) {
	var (
		err error
		t   xml.Token
	)

	xd := xml.NewDecoder(bytes.NewReader(b))
	root := Node{Type: Object, ChildrenNames: map[string]int{}}

	n := &root
	depth := 0
	for t, err = xd.Token(); err == nil; t, err = xd.Token() {
		switch tt := t.(type) {
		case xml.StartElement:
			depth++

			tn := tt.Name.Local
			n.Children = append(n.Children, Node{Name: tn, Parent: n, ChildrenNames: map[string]int{}})

			// Count every name to understand afterwards, what kind of JSON node it is: an object or an array
			if cnt, ok := n.ChildrenNames[tn]; ok {
				n.ChildrenNames[tn] = cnt + 1
			} else {
				n.ChildrenNames[tn] = 1
			}
			n = &n.Children[len(n.Children)-1]

		case xml.EndElement:
			depth--

			if len(n.ChildrenNames) > 0 {
				// If element has at least one child, it's an object or an array
				n.Type = Object

				// If at least one child name repeats, it an array
				for _, cnt := range n.ChildrenNames {
					if cnt > 1 {
						n.Type = Array
						break
					}
				}
			}
			n = n.Parent

		case xml.CharData:
			if depth == 0 {
				continue
			}

			v := strings.TrimSpace(string(tt))
			if v != "" {
				if f, ok := isFloat(v); ok {
					n.Type = Float
					n.ValueFloat = f
				} else if i, ok := isInt(v); ok {
					n.Type = Integer
					n.ValueInt = i
				} else {
					n.Type = String
					n.Value = strings.Trim(v, `"`)
				}
			}
		}
	}

	// Check if root node should be treated as an array
	for _, cnt := range root.ChildrenNames {
		if cnt > 1 {
			root.Type = Array
			break
		}
	}

	r, _ := json.Marshal(tree2map(&root))

	if !errors.Is(err, io.EOF) {
		return nil, err
	}

	return r, nil
}

func tree2map(n *Node) map[string]interface{} {
	m := make(map[string]interface{})

	switch n.Type {
	case String:
		m[n.Name] = n.Value

	case Integer:
		m[n.Name] = n.ValueInt

	case Float:
		m[n.Name] = n.ValueFloat

	case Array:
		s := make(map[string][]interface{}, 0)

		for i := range n.Children {
			c := n.Children[i]
			rc := tree2map(&c)
			switch c.Type {
			case String, Integer, Float:
				s[c.Name] = append(s[c.Name], rc[c.Name])
			default:
				s[c.Name] = append(s[c.Name], rc)
			}
		}

		for k, v := range s {
			if len(v) == 1 {
				m[k] = v[0]
			} else {
				m[k] = v
			}
		}

	case Object:
		for i := range n.Children {
			c := n.Children[i]
			rc := tree2map(&c)

			switch c.Type {
			case String, Integer, Float:
				m[c.Name] = rc[c.Name]
			default:
				m[c.Name] = rc
			}
		}
	}

	return m
}

func isFloat(s string) (float64, bool) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}

	return f, true
}

func isInt(s string) (int64, bool) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}

	return i, true
}
