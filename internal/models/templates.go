package models

// Template represents a single template file by its path.
type Template struct {
	Path string `json:"path"`
}

// Templates is a collection of Template objects.
type Templates []Template
