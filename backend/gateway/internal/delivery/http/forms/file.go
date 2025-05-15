package forms

type MessageAttachmentForm struct {
	MediaURLs []string `json:"media"`
	AudioURLs []string `json:"audio"`
	FileURLs  []string `json:"files"`
}
