package controller

import "github.com/freecloudio/freecloud/manager"

type managerContext struct {
	AuthManager *manager.AuthManager
}

var mc *managerContext

func InitManagerContext(auth *manager.AuthManager) {
	mc = &managerContext{
		AuthManager: auth
	}
}