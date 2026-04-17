package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("email ou senha invalidos")
	ErrInactiveUser       = errors.New("usuario inativo")
	ErrInvalidToken       = errors.New("token invalido")
	ErrExpiredToken       = errors.New("token expirado")
	ErrForbidden          = errors.New("acesso negado para este perfil")
)
