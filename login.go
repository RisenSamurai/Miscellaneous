package server

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"net/http"
	"oacs/server/postgres"
	"path/filepath"
	"strconv"
	"strings"
)

func SignIn(c *gin.Context) {

	var loginReq postgres.Users
	var email string
	var pass string
	var role int
	var uID int

	sql := `SELECT u.id, u.email, u.role , t.password
	FROM users AS u
    LEFT JOIN crypto AS t
    ON u.id = t.user_id 
    WHERE u.email = $1`

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query for user password!", err)
	}

	defer stmt.Close()

	if err := c.BindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err,
		})

		return
	}
	err = stmt.QueryRow(loginReq.Email).Scan(&uID, &email, &role, &pass)
	if err != nil {
		log.Println("Could not scan the result!", err)
	}
	log.Println("Got Email: ", email+" "+loginReq.Email)
	log.Println("Got Password: ", pass)
	log.Println("Got Role: ", role)

	dbEmail := strings.TrimSpace(email)

	if dbEmail == loginReq.Email && pass == loginReq.Password {

		if role == 1 {

			log.Println("Got Email: ", loginReq.Email)
			log.Println("Got Password: ", loginReq.Password)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"isAdmin": true,
			})

			//sUID := strconv.Itoa(uID)

			//CreateQR(uID)
			//SendLogQR(dbEmail, sUID+".qr.png", c)
			log.Println("Over!")

		} else if role != 1 {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"isAdmin": false,
			})
		}

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "isAdmin": false})
	}

}

func LogOut(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"success": true, "isAdmin": false,
	})
}

func GenerateToken() (string, error) {
	bytes := make([]byte, 16) //128bit
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil

}

func SendLogQR(email, QR string, c *gin.Context) {

	var (
		From    string
		Subject string
	)

	qrPath := filepath.Join("public", "qr", QR)

	Subject = "Your QR Code"
	From = "oacsgroup@gmail.com"

	var bodyBuffer bytes.Buffer

	tmpl, err := template.ParseFiles("public/templates/mail_template.html")
	if err != nil {
		log.Println("Error parsing template:", err)
		return
	}

	// Render the template into bodyBuffer
	err = tmpl.Execute(&bodyBuffer, gin.H{
		"title":          Subject,
		"qrCodeFileName": QR, // Make sure this is correctly referenced in your template
	})
	if err != nil {
		log.Println("Error executing template:", err)
		return
	}

	// Prepare the email message
	m := gomail.NewMessage()
	m.SetHeader("From", From)
	m.SetHeader("To", email)
	m.SetHeader("Subject", Subject)
	m.SetBody("text/html", bodyBuffer.String())
	m.Embed(qrPath, gomail.SetHeader(map[string][]string{
		"Content-ID": {"<myimagecid>"},
	}))

	log.Println("Sending....")

	d := gomail.NewDialer("DESKTOP-MC6GJ0P", 25, "", "") //provider port 857

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Could not send the email", err)
		return
	}

}

func CreateQR(userID int) error {
	suID := strconv.Itoa(userID)
	URL := "http://192.168.32.239:3037/sign-in/?token=" + SetToken(userID) + "&id=" + suID // Ensure SetToken handles errors

	err := qrcode.WriteFile(URL, qrcode.Medium, 256, filepath.Join("public", "qr", suID+".qr.png"))
	if err != nil {
		log.Println("Could not encode the QR code!", err)
		return err
	}

	sql := `INSERT INTO qr (user_id, qr_path) VALUES ($1, $2)`

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad QR data insertion", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, filepath.Join("public", "qr", suID+".qr.png"))
	if err != nil {
		log.Println("Error executing insert statement", err)
		return err
	}

	return nil
}

func SetToken(userID int) string {
	token, err := GenerateToken()

	sql := `INSERT INTO token (user_id, token) VALUES ($1, $2)`

	_, err = postgres.GetDB().Exec(sql, userID, token)
	if err != nil {
		log.Println("Bad exec query at SetToken!", err)
	}

	return token

}
