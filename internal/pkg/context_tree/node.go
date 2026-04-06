package context_tree

type Node struct {
	freq     map[byte]int     // сколько раз символ встретился после этого контекста
	children map[string]*Node // контексты длиннее на 1
	total    int              // сумма всех freq
}
