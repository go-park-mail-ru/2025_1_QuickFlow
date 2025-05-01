package models

import (
	"io"
)

type AccessMode int

const (
	AccessPublic AccessMode = iota
	AccessPrivate
)

type File struct {
	Reader     io.Reader
	Name       string
	Size       int64
	Ext        string
	MimeType   string
	AccessMode AccessMode
	URL        string
}

func (f File) String() string {
	return f.Name
}
