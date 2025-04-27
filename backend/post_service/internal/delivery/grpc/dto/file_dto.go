package dto

type AccessModeDTO int32

const (
	AccessPublic  AccessModeDTO = 0
	AccessPrivate AccessModeDTO = 1
)

type FileDTO struct {
	Content    []byte
	Name       string
	Size       int64
	Ext        string
	MimeType   string
	AccessMode AccessModeDTO
	URL        string
}
