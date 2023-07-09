package monk

type MonkBackend struct {
}

func (m *MonkBackend) Log(msg string) {
	println(msg)
}
