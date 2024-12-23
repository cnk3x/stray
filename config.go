package main

type Shortcuts struct {
	Name      string             `json:"name,omitempty" toml:"name,omitempty" ini:"name,omitempty"`
	Args      map[string]string  `json:"args,omitempty" toml:"args,omitempty" ini:"args,omitempty"`
	Shortcuts map[string]Command `json:"shortcuts,omitempty" toml:"shortcuts,omitempty" ini:"shortcuts,omitempty"`
}
