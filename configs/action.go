package configs

type Action int

const (
	Register Action = iota
	Renew
	Cancel
	Delete
)
