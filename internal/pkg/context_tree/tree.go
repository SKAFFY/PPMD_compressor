package context_tree

type ContextTree struct {
	root     *Node
	maxOrder int
}

func NewContextTree(maxOrder int) *ContextTree {
	return &ContextTree{
		root: &Node{
			Freq:     make(map[byte]int),
			Children: make(map[byte]*Node),
			Total:    0,
		},
		maxOrder: maxOrder,
	}
}

// GetNode возвращает узел для заданного контекста (последовательности байт).
// Если контекст не существует, возвращает nil.
func (t *ContextTree) GetNode(context []byte) *Node {
	node := t.root
	for _, c := range context {
		child := node.Children[c]
		if child == nil {
			return nil
		}
		node = child
	}
	return node
}

// Update обновляет статистику для всех суффиксов контекста (от максимального порядка до 0).
// Создаёт недостающие узлы по пути.
func (t *ContextTree) Update(sym byte, context []byte) {
	// Собираем все узлы от корня до узла полного контекста
	nodes := []*Node{t.root}
	current := t.root
	for _, c := range context {
		if current.Children[c] == nil {
			current.Children[c] = &Node{
				Freq:     make(map[byte]int),
				Children: make(map[byte]*Node),
				Total:    0,
			}
		}
		current = current.Children[c]
		nodes = append(nodes, current)
	}
	// Обновляем частоты во всех собранных узлах (суффиксы контекста)
	for _, node := range nodes {
		node.Freq[sym]++
		node.Total++
	}
}
