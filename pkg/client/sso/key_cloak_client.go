package sso

import (
	"github.com/Nerzal/gocloak/v7"
)

type KeyCloakMiddleware struct {
	Keycloak *KeyCloak
}


type KeyCloak struct {
	GoCloak		 gocloak.GoCloak
	ClientId     string
	ClientSecret string
	Realm        string
}


func KeyCloakInit(keycloak *KeyCloak) *KeyCloakMiddleware {
	return &KeyCloakMiddleware{Keycloak: keycloak}
}


