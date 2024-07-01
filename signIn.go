package server

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"oacs/server/lib"
	"oacs/server/postgres"
	"os"
	"time"
)

type Cred struct {
	Email    string `json:"Email"`
	Password string `json:"Password"`
}

func generateJWT(email string, isAdmin bool) (string, error) {

	err := godotenv.Load()
	if err != nil {
		return "Can't read the .env file!", err
	}

	secretKey := os.Getenv("SECRET_KEY")

	if secretKey == "" {
		log.Println("SECRET_KEY is empty, generating a new one!")
		envMap, err := godotenv.Read(".env")
		if err != nil {
			log.Println("Cant read .env file!", err)
			envMap = make(map[string]string)
		}
		key := make([]byte, 32)
		_, err = rand.Read(key)
		if err != nil {
			log.Println("Can't generate a secret key!", err)
			return "", err
		}

		envMap["SECRET_KEY"] = hex.EncodeToString(key)
		err = godotenv.Write(envMap, ".env")
		if err != nil {

			log.Println("Can't write to .env file!", err)
			return "", err

		}
	}

	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["email"] = email
	claims["isAdmin"] = isAdmin
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "Can't assign token!", err
	}

	return tokenString, nil
}

func Login(c *gin.Context) {

	var cred Cred
	dbToken := ""
	role := ""
	email := ""
	var sql string

	log.Println("Got to Login!")

	err := c.BindJSON(&cred)
	if err != nil {
		log.Println("Error with BindJSON!", err)
	}

	log.Println(cred)

	sql = ` SELECT c.password, u.role, u.email  FROM users AS u
            LEFT JOIN crypto AS c ON u.id = c.user_id      
            WHERE email = $1`

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad prepare!", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(cred.Email).Scan(&dbToken, &role, &email)
	if err != nil {
		log.Println("Can't get dbtoken", err)
	}

	isOK := lib.VerifyPassword(dbToken, cred.Password)

	log.Println("ISOK = ", isOK)

	if isOK {
		if role == "1" {

			tokenString, err := generateJWT(cred.Email, true)
			if err != nil {
				log.Println("Can't generate token!", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Can't generate token!",
				})

				return
			}

			c.JSON(http.StatusOK, gin.H{
				"token":   tokenString,
				"isAdmin": true,
			})

		}
		if role == "2" {
			tokenString, err := generateJWT(cred.Email, false)
			if err != nil {
				log.Println("Can't generate token!", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Could not generate token!",
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"token":   tokenString,
				"isAdmin": "false",
			})

		}

	} else {

		log.Println(isOK)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Access denied!",
		})
	}

}
