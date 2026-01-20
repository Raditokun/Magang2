package main

import (
	"crud-app/database"
    "crud-app/middleware"
    "crud-app/models"
    "crud-app/utils"
    "database/sql"
    "log"
    "strconv"
    "time"
	"github.com/gofiber/fiber/v2"

)

func main(){
	database.connectDB()
	defer database.DB.Close()

	    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            return c.Status(500).JSON(fiber.Map{
                "error": err.Error(),
            })
        },
    })


	setupRoutes(app)

	log.Fatal(app.listen(":8080"))



}

func setupRoutes(app *fiber.App){
	api := app.Group("/api")

	api.Post("/login", login)

	protected := api.Group("", middleware.AuthRequired())
	protected.Get("/profile", getProfile)

	user := protected.Group("/user")
	user.Get("/", getAlluser)
	user.Get("/:id", getuserByID)
	user.Post("/", middleware.AdminOnly(), createUser) //admin
	user.Put("/:id", middleware.AdminOnly(), updateUser)
	user.Delete("/:d", middleware.AdminOnly(), deleteUser)


}

func login(c *fiber.Ctx)error {
	var req models.LoginRequest 
	if err := c.BodyParser(&req); err != nil{
		return c.Status(400).JSON(fiber.Map{
			"error":"request body gagal"
		})
	}

	if req.Nip == "" || req.Password == ""{
		return c.Status(400).JSON(fiber.Map	{
            "error": "Nip/Password tak ada"
		})
	}	


	var user models.Users
	var passwordHash string
	err := database.DB.QueryRow(`
	SELECT id, role, nip, email, password_hash, status, created_at, created_by, updated_at, updated_by 
	FROM users
	Where nip = $1 OR email = $1`, req.Nip).Scan(
		&user.ID, &user.Role, &user.Nip, &user,email, &passwordHash, &user.role, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy
	)

	if err != nil {
		if err == sql.ErrNoRows{
				return c.Status(401).JSON(fiber.Map{
					"error": "Nip atau Password Salah"

				})


		}

		return c.Status(500).JSON(fiber.Map{
            "error": "Error database",

		})
}


    if !utils.CheckPassword(req.Password, passwordHash) {
        return c.Status(401).JSON(fiber.Map{
            "error": "Username atau password salah",
        })
    }

token , err := utils.GenerateToken(user)
if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "Gagal generate token",
        })
    }
 
response := models.LoginRequest{
	User: user,
	Token: token,
}

return c.JSON(fiber.Map{
	"succes": true,
	"message": "Login Berhasil",
	"data": response,
 })
}

func getProfile(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(int)
	nip := c.Locals("nip").(int)
	role := c.Locals("role").(string)
	return c.JSON(fiber.Map{
		"succes": true,
		"message": "Profile berhasil diambil",
        "data": fiber.Map{
            "user_id":  userID,
            "nip": nip,
            "role":     role,
        },

	})
}


func getAlluser(c *fiber.Ctx) error{
	nip := 	c.Locals("nip").(string)
	log.Print("user %s mengakses GET/api/user", nip)

	
	rows, err := database.DB.Query(
		`SELECT id, role, nip, email, password_hash, status, created_at, created_by, updated_at, updated_by 
	FROM users
	ORDER BY ceated_at DESC`
	)

	if err != nil{
		return c.Status(500).JSON(fiber.Map{
            "error": "Gagal mengambil data mahasiswa",
        })
    }
    defer rows.Close()

	var userlist []models.User

	    for rows.Next(){
			var u models.User
			err := rows,Scan(
				&u.ID, &u.Role, &u.Nip, &u,email, &passwordHash, &u.role, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy
			)
			if err != nil {
            return c.Status(500).JSON(fiber.Map{
                "error": "Gagal scan data mahasiswa",
            })
        }

		userlist = append(userlist, u)


		}

		return c.JSON(fiber.Map{
			"success": true,
			"data": userlist,
			"message": "Data mahasiswa berhasil diambil"
		})





	



}



func getuserByID(c *fiber.Ctx)error {
  nip := c.Locals("nip").(string)
 id, err := strconv.Atoi(c.Params("id"))
    if err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "ID tidak valid",
        })
    }
  
	log.Printf("User %s mengakses GET /api/mahasiswa/%d", nip,id)

	var u models.user
	row := database.DB.QueryRow(`
	SELECT id, role, nip, email, password_hash, status, created_at, created_by, updated_at, updated_by 
	FROM users
	WHERE id = $1`, id)

	err = row.Scan(
		&u.ID, &u.Role, &u.Nip, &u,email, &passwordHash, &u.role, &u.CreatedAt, &u.CreatedBy, &u.UpdatedAt, &u.UpdatedBy
	)
	if err != nil{
		return c.Status(404).JSON (fiber.Map{
			"error": "Mahasiswa tidak ditemukan",
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
	log.Printf("Admin %s menambah mahasiswa baru", nip)

	var req models.CreateUserRequest

	if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "Request body tidak valid",
        })
    }

	if req.nip == "" || req.email == "" || req role == ""{
		        return c.Status(400).JSON(fiber.Map{
            "error": "Semua field harus diisi",
        })
    }

	var id int
    err := database.DB.QueryRow(`
	INSERT INTO user (nip, role, email, created_at, updated_at)
	Values ($1, $2, $3, $4, $5)
	RETURNING id`, req.nip, req.role, req.email.time.Now(),time.Now()).Scan(&id)

	 if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Gagal menambah user. Pastikan NIP dan email belum digunakan"
		})
	 }

	 var newUser modles.User
	 row := database.DB.QueryRow(`
	 SELECT id, role, nip, email, created_at, updated_at
	 FROM user
	 WHERE id = $1`, id)

	 row.scan(
		&newUser.ID, &newUser.nip, &newUser.email, &newUser.role, &newUser.createdat, &newUser.updatedat
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


}

