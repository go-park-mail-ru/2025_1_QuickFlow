package forms

type ProfileFriendsInfoOut struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	AvatarURL  string `json:"avatar_url"`
	FirstName  string `json:"firstname"`
	LastName   string `json:"lastname"`
	University string `json:"univ_name,omitempty"`
}

type FriendsOut struct {
	Friends map[ProfileFriendsInfoOut]string `json:"friends"`
}
