package edgedb

func inMap[T comparable, U any](k T, m map[T]U) bool {
	_, ok := m[k]
	return ok
}
