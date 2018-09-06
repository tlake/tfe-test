package backend

type Backend interface {
	GetEntries() ([]string, error)
	AddEntry(string) error
	ClearEntries() error
}
