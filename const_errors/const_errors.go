package consterrors

import "errors"

var ErrEmptyPassword error = errors.New("empty password")
var ErrPasswordToShort error = errors.New("password shorter than 10 characters")
var ErrPasswordMissingElements error = errors.New("password must containt lower case, upper case, digit and at least a symbol")
var ErrEmptyEmail error = errors.New("no email supplied")
var ErrEmptyToken error = errors.New("no token supplied")
var ErrInvalidCredentials error = errors.New("invalid credentials")
var ErrAlreadyConfirmed error = errors.New("already confirmed this request")
var ErrInvalidToken error = errors.New("invalid token")
var ErrNoToken error = errors.New("no token")
var ErrNoGitToken error = errors.New("no git token")
var ErrNotLoggedIn error = errors.New("not logged in")
