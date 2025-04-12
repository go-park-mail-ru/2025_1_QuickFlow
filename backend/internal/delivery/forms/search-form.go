package forms

import (
	"errors"
	"net/url"
	"strconv"
)

type SearchForm struct {
	Username   string `json:"username" validate:"required"`
	UsersCount uint   `json:"users_count" validate:"required"`
}

func (s *SearchForm) Unpack(values url.Values) error {
	if !values.Has("username") {
		return errors.New("username parameter missing")
	}

	if !values.Has("users_count") {
		return errors.New("users_count parameter missing")
	}

	s.Username = values.Get("username")

	usersCount, err := strconv.Atoi(values.Get("users_count"))
	if err != nil {
		return errors.New("failed to parse users_count")
	}

	s.UsersCount = uint(usersCount)

	return nil
}
