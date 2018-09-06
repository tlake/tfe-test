package backend

type MemoryBackend struct {
	entries []string
}

func NewMemoryBackend() *MemoryBackend {
	return &MemoryBackend{
		entries: []string{},
	}
}

func (m *MemoryBackend) GetEntries() ([]string, error) {
	return m.entries, nil
}

func (m *MemoryBackend) AddEntry(message string) error {
	m.entries = append(m.entries, message)
	return nil
}

func (m *MemoryBackend) ClearEntries() error {
	m.entries = []string{}
	return nil
}
