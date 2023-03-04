package xml2json

const (
	String = iota
	Integer
	Float
	Array
	Object
)

type Node struct {
	Type          int
	Name          string
	Value         string
	ValueInt      int64
	ValueFloat    float64
	Parent        *Node
	Children      []Node
	ChildrenNames map[string]int
}
