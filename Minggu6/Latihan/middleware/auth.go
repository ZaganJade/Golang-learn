package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"

	"backend-golang/Minggu6/Latihan/helper"
)

// RequireAuth memeriksa access token pada header Authorization.
// Bila tokennya sah, identitas pemakai disimpan di Locals agar dapat
// dibaca service tanpa memeriksa ulang.
func RequireAuth(jwtManager *helper.JWTManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		token, err := bearerToken(c)
		if err != nil {
			// WWW-Authenticate adalah header baku yang menyertai 401.
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			return helper.Fail(c, fiber.StatusUnauthorized,
				"header Authorization tidak ada atau salah bentuk")
		}

		authUser, err := jwtManager.Parse(token)
		if err != nil {
			c.Set("WWW-Authenticate", `Bearer realm="api"`)
			// Membedakan "kedaluwarsa" dari "tidak valid" aman dilakukan:
			// client memang perlu tahu kapan harus memanggil /auth/refresh.
			if errors.Is(err, helper.ErrExpiredToken) {
				return helper.Fail(c, fiber.StatusUnauthorized, "access token kedaluwarsa")
			}
			return helper.Fail(c, fiber.StatusUnauthorized, "access token tidak valid")
		}

		c.Locals(helper.LocalsAuthUser, authUser)
		return c.Next()
	}
}

func bearerToken(c fiber.Ctx) (string, error) {
	header := c.Get(fiber.HeaderAuthorization)
	if header == "" {
		return "", errors.New("header kosong")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", errors.New("format bukan Bearer")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", errors.New("token kosong")
	}
	return token, nil
}

// LoginRateLimiter membatasi jumlah percobaan login dari satu alamat IP.
// Tanpa pembatasan ini, penyerang dapat mencoba ribuan password per menit
// (serangan brute force) tanpa hambatan apa pun.
func LoginRateLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		KeyGenerator: func(c fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c fiber.Ctx) error {
			c.Set("Retry-After", "60")
			return helper.Fail(c, fiber.StatusTooManyRequests,
				"terlalu banyak percobaan login, coba lagi dalam satu menit")
		},
	})
}

// RequirePermission menolak request yang role-nya tidak memiliki
// permission tertentu. Dipasang pada route yang haknya dapat diputuskan
// TANPA melihat isi data --- misalnya "boleh melihat daftar seluruh user".
//
// Untuk keputusan yang bergantung pada isi data (misalnya "boleh mengubah
// data ini karena miliknya sendiri"), pemeriksaan tidak bisa di sini:
// middleware belum tahu data siapa yang akan disentuh. Pemeriksaan
// semacam itu berada di layer service.
func RequirePermission(perms *helper.PermissionSet, permission string) fiber.Handler {
	return func(c fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			// Sampai di sini tanpa identitas berarti RequireAuth belum
			// dipasang. Tolak, jangan diloloskan.
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if !perms.Can(user.Role, permission) {
			return helper.Fail(c, fiber.StatusForbidden,
				"role "+user.Role+" tidak memiliki hak "+permission)
		}
		return c.Next()
	}
}

// RequireRole memeriksa nama role secara langsung.
//
// Cara ini lebih sederhana, tetapi lebih kaku: menambah satu role baru
// berarti menyunting seluruh route yang menyebut nama role lama.
// Disediakan di sini sebagai pembanding, bukan sebagai pilihan utama.
func RequireRole(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c fiber.Ctx) error {
		user, ok := helper.CurrentUser(c)
		if !ok {
			return helper.Fail(c, fiber.StatusUnauthorized, "belum terautentikasi")
		}
		if _, granted := allowed[user.Role]; !granted {
			return helper.Fail(c, fiber.StatusForbidden,
				"role Anda tidak berhak mengakses endpoint ini")
		}
		return c.Next()
	}
}
