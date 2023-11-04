package controllers

import (
	"example/RESTful-API-login/initializers"
	"example/RESTful-API-login/models"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

func Signup(c *gin.Context) {
	// Get the email/pass off req body
	var body struct {
		Fname    string
		Lname    string
		Phone    string
		Email    string `gorm:"unique"`
		Password string
		Adm      bool
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	// Hash the password
	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password), 10)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})

		return
	}

	// Create the user
	// user := models.User{Email: body.Email, Password: string(hash)}
	user := models.User{
		Fname:    body.Fname,
		Lname:    body.Lname,
		Phone:    body.Phone,
		Email:    body.Email,
		Password: string(hash),
		Adm:      false,
	}
	result := initializers.DB.Create(&user)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create user",
		})

		return
	}

	// Responde
	c.JSON(http.StatusOK, gin.H{})
}

func Login(c *gin.Context) {
	// Get the email and pass off req body
	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	// Look up requested user
	var user models.User
	initializers.DB.First(&user, "email = ?", body.Email)

	if user.ID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})

		return
	}

	// Compare sent in pass with saved user pass hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(body.Password))

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})

		return
	}

	// Generate a jwt token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userId":         user.ID, // subject = userd ID
		"expirationDate": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})

		return
	}

	// Send it back
	//The SameSite attribute of the Set-Cookie HTTP response header allows you to declare if your cookie should be restricted
	//to a first-party or same-site context.

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})
}

func Validate(c *gin.Context) {
	user, _ := c.Get("user")

	c.JSON(http.StatusOK, gin.H{
		"message": user,
	})
}

func Update(c *gin.Context) {
	var requestDTO = struct {
		// gorm.Model
		ID       uint   `json:"id"`
		Fname    string `json:"fname,omitempty"`
		Lname    string `json:"lname,omitempty"`
		Phone    string `json:"phone,omitempty"`
		Email    string `json:"email,omitempty" gorm:"unique"`
		Password string `json:"password,omitempty"`
		Adm      bool   `json:"adm,omitempty" gorm:"default:true"`
	}{}

	if c.Bind(&requestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})

		return
	}

	// Check credentials

	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	})

	var requestOwner models.User
	initializers.DB.Take(&requestOwner, token.Claims.(jwt.MapClaims)["userId"])

	if !requestOwner.Adm && requestDTO.ID != requestOwner.ID {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	// Map to store the fields to be changed (only change fields that are not empty)
	var mp map[string]any = map[string]any{}
	if len(requestDTO.Fname) > 0 {
		mp["fname"] = requestDTO.Fname
	}
	if len(requestDTO.Lname) > 0 {
		mp["lname"] = requestDTO.Lname
	}
	if len(requestDTO.Phone) > 0 {
		mp["phone"] = requestDTO.Phone
	}
	if len(requestDTO.Email) > 0 {
		mp["email"] = requestDTO.Email
	}
	if len(requestDTO.Password) > 0 {
		mp["password"] = requestDTO.Password
	}

	// Verify if requester is Adm
	if requestOwner.Adm && requestOwner.ID != requestDTO.ID {
		mp["adm"] = requestDTO.Adm
	}

	// Update the fields that have changed
	var conn = initializers.DB.Debug().Table("users").Where("id = ?", requestDTO.ID).Updates(mp)
	if conn.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update fields",
			"error":   conn.Error,
		})
	}

	// Get users from DB to check its status after update
	var usr models.User = models.User{}
	conn = initializers.DB.Where("id = ?", requestDTO.ID).First(&usr)
	if conn.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to get user info from database.",
			"error":   conn.Error,
		})
	}

	// Send it back
	c.JSON(http.StatusOK, gin.H{
		"message": usr,
		"map":     mp,
	})
}

func ListUsers(c *gin.Context) {

	//Check credentials - Only ADM can list all users

	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
		})
		return
	}
	token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(os.Getenv("SECRET")), nil
	})

	var requestOwner models.User
	initializers.DB.Take(&requestOwner, token.Claims.(jwt.MapClaims)["userId"])

	//If not ADM return unauthorized access msg

	if !requestOwner.Adm {
		c.JSON(http.StatusForbidden, gin.H{})
		return
	}

	//If ADM return list of users
	var users []models.User
	result := initializers.DB.Find(&users)

	if result != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": users,
		})
	}
}
