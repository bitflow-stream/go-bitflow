package script

import (
	"fmt"
)

type GenericStack [][]interface{}

func (s *GenericStack) Clear() {
	*s = nil
}

func (s *GenericStack) Push(objs ...interface{}) {
	*s = append(*s, objs)
}

func (s *GenericStack) Pop() []interface{} {
	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return res
}

func (s *GenericStack) Append(objs ...interface{}) {
	i := len(*s) - 1
	(*s)[i] = append((*s)[i], objs...)
}

func (s *GenericStack) Peek() []interface{} {
	return (*s)[len(*s)-1]
}

func (s *GenericStack) PeekSingle() interface{} {
	return s.assertSingle(s.Peek())
}

func (s *GenericStack) PopSingle() interface{} {
	return s.assertSingle(s.Pop())
}

func (s *GenericStack) assertSingle(objs []interface{}) interface{} {
	if len(objs) != 1 {
		panic(fmt.Errorf("Top-of-stack contains %v objects instead of 1", len(objs)))
	}
	return objs[0]
}
