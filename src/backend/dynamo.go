package backend

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
	"time"
)

type Entry struct {
	ID      string `dynamo:"id"`
	Message string `dynamo:"message"`
}

type DynamoBackend struct {
	table dynamo.Table
}

func NewDynamoBackend(table string) *DynamoBackend {
	config := defaults.Get().Config
	db := dynamo.New(session.New(config), nil)

	return &DynamoBackend{
		table: db.Table(table),
	}
}

func (d *DynamoBackend) getEntries() ([]Entry, error) {
	var entries []Entry
	if err := d.table.Scan().All(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}

func (d *DynamoBackend) GetEntries() ([]string, error) {
	entries, err := d.getEntries()
	if err != nil {
		return nil, err
	}

	messages := make([]string, len(entries))
	for i, e := range entries {
		messages[i] = e.Message
	}

	return messages, nil
}

func (d *DynamoBackend) AddEntry(message string) error {
	entry := Entry{
		ID:      time.Now().Format(time.UnixDate),
		Message: message,
	}

	return d.table.Put(entry).Run()
}

func (d *DynamoBackend) ClearEntries() error {
	entries, err := d.getEntries()
	if err != nil {
		return err
	}

	for _, e := range entries {
		if err := d.table.Delete("id", e.ID).Run(); err != nil {
			return err
		}
	}

	return nil
}
