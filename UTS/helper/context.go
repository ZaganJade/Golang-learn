package helper

import (
	"github.com/gofiber/fiber/v3"

	"backend-golang/UTS/app/model"
)

const LocalsAuthUser = "authUser"

// CurrentUser mengambil isi token yang dipasang middleware RequireAuth.
func CurrentUser(c fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
