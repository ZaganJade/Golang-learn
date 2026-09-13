package helper

import (
	"github.com/gofiber/fiber/v3"

	"backend-golang/Minggu5/Latihan/app/model"
)

// LocalsAuthUser adalah kunci penyimpanan identitas pemakai di dalam
// context request. Dibuat sebagai konstanta agar tidak ada salah ketik
// antara tempat menyimpan dan tempat membaca.
const LocalsAuthUser = "authUser"

func CurrentUser(c fiber.Ctx) (model.AuthUser, bool) {
	user, ok := c.Locals(LocalsAuthUser).(model.AuthUser)
	return user, ok
}
