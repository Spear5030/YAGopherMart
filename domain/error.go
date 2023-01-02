package domain

import "errors"

var ErrUserExists = errors.New("user exists")
var ErrInvalidPassword = errors.New("invalid login or password")
