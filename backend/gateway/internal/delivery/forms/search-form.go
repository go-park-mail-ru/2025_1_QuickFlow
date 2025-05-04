package forms

import (
	"errors"
	"net/url"
	"strconv"
)

type SearchForm struct {
	ToSearch string `json:"string" validate:"required"`
	Count    uint   `json:"count" validate:"required"`
}

func (s *SearchForm) Unpack(values url.Values) error {
	if !values.Has("string") {
		return errors.New("username parameter missing")
	}

	if !values.Has("count") {
		return errors.New("count parameter missing")
	}

	s.ToSearch = values.Get("string")

	usersCount, err := strconv.Atoi(values.Get("count"))
	if err != nil {
		return errors.New("failed to parse count")
	}

	s.Count = uint(usersCount)

	return nil
}
