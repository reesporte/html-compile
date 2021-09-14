package main

// Set is a basic set implementation
type Set struct {
	set map[interface{}]bool
}

// NewSet initializes an empty interface{} set
func NewSet() *Set {
	s := &Set{}
	s.set = make(map[interface{}]bool)
	return s
}

// Add adds a value to the set
func (s *Set) Add(val interface{}) {
	s.set[val] = true
}

// Delete removes a value from the set
func (s *Set) Delete(val interface{}) {
	delete(s.set, val)
}

// Contains returns whether the value is in the set
func (s *Set) Contains(val interface{}) bool {
	_, contains := s.set[val]
	return contains
}

// Update adds a list of values to the set
func (s *Set) Update(vals []interface{}) {
	for _, val := range vals {
		s.set[val] = true
	}
}

// ShallowCopy returns a shallow copy of the Set
func (s *Set) ShallowCopy() *Set {
	copiedSet := NewSet()

	for k := range s.set {
		copiedSet.Add(k)
	}

	return copiedSet
}

// Values returns a slice of interface{}s of the values stored in the Set
func (s *Set) Values() []interface{} {
	i := 0
	values := make([]interface{}, len(s.set))
	for k := range s.set {
		values[i] = k
		i++
	}
	return values
}

// Stack is a basic stack implementation
type Stack []interface{}

// IsEmpty returns whether the stack is empty
func (s *Stack) IsEmpty() bool {
	return len(*s) == 0
}

// Push adds val to the stack
func (s *Stack) Push(val interface{}) {
	*s = append(*s, val)
}

// Pop removes an element from the stack and returns it, as well as whether the stack is empty
func (s *Stack) Pop() interface{} {
	if s.IsEmpty() {
		return nil
	}
	element := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return element
}

// Peek returns the top element of the stack but doesn't remove it
func (s *Stack) Peek() interface{} {
	if s.IsEmpty() {
		return nil
	}
	return (*s)[len(*s)-1]
}
