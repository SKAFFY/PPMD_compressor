package context_tree

type ContextTree struct {
	root          *Node
	maxOrder      int
	updateHistory []byte
}

func NewContextTree(maxOrder int) *ContextTree {
	return &ContextTree{}
}
