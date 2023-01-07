package domain

import "errors"

var ErrUserExists = errors.New("user exists")
var ErrInvalidPassword = errors.New("invalid login or password")
var ErrAlreadyUploaded = errors.New("the order number has already been uploaded by this user")
var ErrAnotherUser = errors.New("the order number has already been uploaded by another user")
var ErrInsufficientBalance = errors.New("the account had insufficient balance")
