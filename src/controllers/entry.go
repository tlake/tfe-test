package controllers

import (
	"github.com/quintilesims/guides/guestbook/backend"
	"github.com/zpatrick/fireball"
)

type EntryController struct {
	backend backend.Backend
}

func NewEntryController(b backend.Backend) *EntryController {
	return &EntryController{
		backend: b,
	}
}

func (m *EntryController) Routes() []*fireball.Route {
	routes := []*fireball.Route{
		{
			Path: "/",
			Handlers: fireball.Handlers{
				"GET":  m.ListEntries,
				"POST": m.AddEntry,
			},
		},
		{
			Path: "/clear",
			Handlers: fireball.Handlers{
				"POST": m.ClearEntries,
			},
		},
	}

	return routes
}

func (m *EntryController) ListEntries(c *fireball.Context) (fireball.Response, error) {
	entries, err := m.backend.GetEntries()
	if err != nil {
		return nil, err
	}

	return c.HTML(200, "index.html", entries)
}

func (m *EntryController) AddEntry(c *fireball.Context) (fireball.Response, error) {
	entry := c.Request.FormValue("entry")
	if err := m.backend.AddEntry(entry); err != nil {
		return nil, err
	}

	return fireball.Redirect(301, "/"), nil
}

func (m *EntryController) ClearEntries(c *fireball.Context) (fireball.Response, error) {
	if err := m.backend.ClearEntries(); err != nil {
		return nil, err
	}

	return fireball.Redirect(301, "/"), nil
}
