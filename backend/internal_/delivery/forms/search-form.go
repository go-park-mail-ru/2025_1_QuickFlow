package forms

import (
	"errors"
	"net/url"
	"strconv"
)

type SearchForm struct {
	ToSearch   string `json:"string" validate:"required"`
	UsersCount uint   `json:"users_count" validate:"required"`
}

func (s *SearchForm) Unpack(values url.Values) error {
	if !values.Has("string") {
		return errors.New("username parameter missing")
	}

	if !values.Has("users_count") {
		return errors.New("users_count parameter missing")
	}

	s.ToSearch = values.Get("string")

	usersCount, err := strconv.Atoi(values.Get("users_count"))
	if err != nil {
		return errors.New("failed to parse users_count")
	}

	s.UsersCount = uint(usersCount)

	return nil
}
