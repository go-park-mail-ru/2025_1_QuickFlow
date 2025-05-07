package dto

type AccessModeDTO int32

const (
	AccessPublic  AccessModeDTO = 0
	AccessPrivate AccessModeDTO = 1
)

type UploadFileDTO struct {
	Content    []byte
	Name       string
	Size       int64
	Ext        string
	MimeType   string
	AccessMode AccessModeDTO
	URL        string
}

type UploadFileResponseDTO struct {
	FileURL string
}

type UploadManyFilesDTO struct {
	Files []*UploadFileDTO
}

type UploadManyFilesResponseDTO struct {
	FileURLs []string
}

type GetFileURLDTO struct {
	FileName string
}

type GetFileURLResponseDTO struct {
	URL string
}

type DeleteFileDTO struct {
	FileURL string
}
