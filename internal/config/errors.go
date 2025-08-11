package config

import "errors"

// Errores de configuración
var (
	ErrMissingAccessToken = errors.New("access token de Mercado Pago es requerido")
	ErrInvalidCredentials = errors.New("credenciales de Mercado Pago inválidas")
	ErrSDKInitialization  = errors.New("error al inicializar el SDK de Mercado Pago")
)
