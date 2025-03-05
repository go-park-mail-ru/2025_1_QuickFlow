package forms

type PostForm struct {
	Desc string   `json:"text"`
	Pics []string `json:"pics"`
}
