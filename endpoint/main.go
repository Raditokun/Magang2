package main

import (
	"crud-app/database"
	"crud-app/middleware"
	"crud-app/models"
	"crud-app/utils"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

func main() {
	database.ConnectDB()
	defer database.DB.Close()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	setupRoutes(app)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3002"
	}
	log.Printf("Server running on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App) {
	api := app.Group("/api")

	api.Post("/login", login)

	// DEBUG: Generate password hash (remove in production)
	api.Get("/hash/:password", func(c *fiber.Ctx) error {
		password := c.Params("password")
		hash, err := utils.HashPassword(password)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"password": password,
			"hash":     hash,
		})
	})
	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/profile", getProfile)

	user := protected.Group("/user")
	user.Get("/", getAlluser)
	user.Get("/:id", getuserByID)
	user.Post("/", middleware.AdminOnly(), createUser)
	user.Put("/:id", middleware.AdminOnly(), updateUser)
	user.Delete("/:id", middleware.AdminOnly(), deleteUser)
}

func login(c *fiber.Ctx) error {
	log.Printf("DEBUG: Raw body: %q", string(c.Body()))
	log.Printf("DEBUG: Content-Type header: %q", c.Get("Content-Type"))

	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("DEBUG: BodyParser error: %v", err)
		return c.Status(400).JSON(fiber.Map{
			"error":   "request body gagal",
			"details": err.Error(),
		})
	}

	log.Printf("DEBUG: Parsed req: %+v", req)

	if req.Nip == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Nip/Password tak ada",
		})
	}

	var user models.Users
	var passwordHash string
	err := database.DB.QueryRow(`
	SELECT id, role, nip, email, password_hash, status, created_at, created_by, updated_at, updated_by 
	FROM users
	WHERE nip = $1 OR email = $1`, req.Nip).Scan(
		&user.ID, &user.Role, &user.Nip, &user.Email, &passwordHash, &user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(401).JSON(fiber.Map{
				"error": "Nip atau Password Salah",
			})
		}

		return c.Status(500).JSON(fiber.Map{
			"error": "Error database",
		})
	}

	log.Printf("DEBUG: Found user with NIP: %s, Hash from DB: %s", user.Nip, passwordHash)
	log.Printf("DEBUG: Password from request: %s", req.Password)

	if !utils.CheckPasswordHash(req.Password, passwordHash) {
		log.Printf("DEBUG: Password check FAILED")
		return c.Status(401).JSON(fiber.Map{
			"error": "Username atau password salah",
		})
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal generate token",
		})
	}

	response := models.LoginResponse{
		User:  user,
		Token: token,
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login Berhasil",
		"data":    response,
	})
}

func getProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	nip := c.Locals("nip").(string)
	role := c.Locals("role").(string)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile berhasil diambil",
		"data": fiber.Map{
			"user_id": userID,
			"nip":     nip,
			"role":    role,
		},
	})
}

func getAlluser(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	log.Printf("user %s mengakses GET/api/user", nip)

	rows, err := database.DB.Query(
		`SELECT id, role, nip, email, status, created_at, created_by, updated_at, updated_by 
	FROM users
	ORDER BY created_at DESC`,
	)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal mengambil data user",
		})
	}
	defer rows.Close()

	var userlist []models.User

	for rows.Next() {
		var u models.User
		err := rows.Scan(
			&u.ID, &u.Role, &u.Nip, &u.Email, &u.Status, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy,
		)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Gagal scan data user",
			})
		}

		userlist = append(userlist, u)
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    userlist,
		"message": "Data user berhasil diambil",
	})
}

func getuserByID(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	log.Printf("User %s mengakses GET /api/user/%d", nip, id)

	var u models.User
	row := database.DB.QueryRow(`
	SELECT id, role, nip, email, status, created_at, created_by, updated_at, updated_by 
	FROM users
	WHERE id = $1`, id)

	err = row.Scan(
		&u.ID, &u.Role, &u.Nip, &u.Email, &u.Status, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy,
	)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "user tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    u,
		"message": "Data user berhasil diambil",
	})
}

func createUser(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	log.Printf("Admin %s menambah user baru", nip)

	var req models.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Request body tidak valid",
		})
	}

	if req.Nip == "" || req.Email == "" || req.Role == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Semua field harus diisi",
		})
	}

	var id int
	err := database.DB.QueryRow(`
	INSERT INTO users (nip, role, email, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5)
	RETURNING id`, req.Nip, req.Role, req.Email, time.Now(), time.Now()).Scan(&id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menambah user. Pastikan NIP dan email belum digunakan",
		})
	}

	var newUser models.User
	row := database.DB.QueryRow(`
	 SELECT id, role, nip, email, status, created_at, created_by, updated_at, updated_by
	 FROM users
	 WHERE id = $1`, id)

	row.Scan(
		&newUser.ID, &newUser.Role, &newUser.Nip, &newUser.Email, &newUser.Status, &newUser.CreatedAt, &newUser.CreatedBy, &newUser.UpdatedAt, &newUser.UpdatedBy,
	)

	return c.Status(201).JSON(fiber.Map{
		"success": true,
		"data":    newUser,
		"message": "User berhasil ditambahkan",
	})
}

func updateUser(c *fiber.Ctx) error {
	nip := c.Locals("nip").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	log.Printf("Admin %s mengupdate User ID %d", nip, id)

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "body invalid",
		})
	}

	if req.Nip == "" || req.Email == "" || req.Role == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "nip, email dan role harus diisi",
		})
	}

	result, err := database.DB.Exec(
		`UPDATE users
			SET nip = $1, role = $2, email = $3,  updated_at = $4
			WHERE id = $5`,
		req.Nip, req.Role, req.Email, time.Now(), id)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "gagal mengupdate user",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "user tidak ditemukan",
		})
	}

	var updatedUser models.User
	row := database.DB.QueryRow(
		`SELECT id, role, nip, email, status, created_at, created_by, updated_at, updated_by 
	FROM users
	WHERE id = $1`, id,
	)

	row.Scan(
		&updatedUser.ID, &updatedUser.Role, &updatedUser.Nip, &updatedUser.Email, &updatedUser.Status, &updatedUser.CreatedAt, &updatedUser.CreatedBy, &updatedUser.UpdatedAt, &updatedUser.UpdatedBy,
	)
	return c.JSON(fiber.Map{
		"success": true,
		"data":    updatedUser,
		"message": "User berhasil di update",
	})
}

func deleteUser(c *fiber.Ctx) error {
	adminNip := c.Locals("nip").(string)
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	log.Printf("Admin %s menghapus User ID %d", adminNip, id)

	var exists bool
	err = database.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id).Scan(&exists)
	if err != nil || !exists {
		return c.Status(404).JSON(fiber.Map{
			"error": "User tidak ditemukan",
		})
	}

	_, err = database.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		log.Printf("Delete error: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menghapus user",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "User berhasil dihapus",
	})
}
