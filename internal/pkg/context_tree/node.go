package context_tree

type Node struct {
	Freq     map[byte]int   // частота символов после данного контекста
	Children map[byte]*Node // переходы по следующему символу (расширение контекста)
	Total    int            // сумма всех Freq
}
