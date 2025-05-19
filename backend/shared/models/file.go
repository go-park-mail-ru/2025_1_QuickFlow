package models

import (
	"io"
)

type DisplayType string
type AccessMode int

const (
	AccessPublic AccessMode = iota
	AccessPrivate

	DisplayTypeMedia = "media"
	DisplayTypeAudio = "audio"
	DisplayTypeFile  = "file"
)

type File struct {
	Reader      io.Reader
	Name        string
	Size        int64
	Ext         string
	MimeType    string
	AccessMode  AccessMode
	URL         string
	DisplayType DisplayType
}

func (f File) String() string {
	return f.Name
}
