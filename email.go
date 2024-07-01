package server

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"oacs/server/postgres"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type Letter struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func SendOneMail(c *gin.Context) {

	var l Letter
	var message string

	c.BindJSON(&l)

	log.Println(l.From)
	log.Println(l.To)

	if l.From == "" {
		l.From = "oacs.support@gmail.com"
	}

	t, err := template.ParseFiles("public/templates/simple_mail.html")
	log.Println("Parsing txt template")
	if err != nil {
		log.Println("Could not parse email template", err)
		return
	}

	var bodyBuffer bytes.Buffer

	if err := t.Execute(&bodyBuffer, gin.H{
		"title": "OACS SERVICE",
		"body":  l.Body,
	}); err != nil {
		log.Println("Could not execute template", err)
		return
	}

	log.Println("Assigning fields")
	m := gomail.NewMessage()
	m.SetHeader("From", l.From)
	m.SetHeader("To", l.To)
	m.SetHeader("Subject", l.Subject)
	m.SetBody("text/html", bodyBuffer.String())

	d := gomail.NewDialer("DESKTOP-MC6GJ0P", 25, "", "") //provider port 857

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Println("Could not send the email", err)
		message = "Failed to send message!"
		c.JSON(http.StatusBadGateway, gin.H{
			"message": message,
		})
		return
	}

	message = "Message was successfully sent!"

	c.JSON(http.StatusOK, gin.H{
		"message": message,
	})

}

func SendCSVEmail(c *gin.Context) {

	const maxFileSize = 10 << 20

	var user_list []postgres.User_info

	log.Println("Got here!")

	err := c.Request.ParseMultipartForm(maxFileSize)
	if err != nil {
		c.String(http.StatusBadRequest, "File too large!")
		return
	}

	fileHeader, err := c.FormFile("attachment")
	if err != nil {
		log.Println("no file received")
		c.String(http.StatusBadRequest, "No file received!")
		return
	}

	if fileHeader.Size > maxFileSize {
		log.Println("File too large")
		c.String(http.StatusBadRequest, "File is too large!")
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		log.Println("can't read a file!")
		c.String(http.StatusBadRequest, "Can't read the file!")
		return
	}

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil {
		log.Println("Can't read csv file!")
		return
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		var u postgres.User_info

		if err != nil {
			fmt.Println("Error reading row:", err)
			continue // Skip this row and move to the next
		}

		// Assuming the email is the first column
		email := record[0]
		name := record[1]
		surname := record[2]
		city := record[3]
		age := record[4]
		content := record[5]

		u.Email = email
		u.Name = name
		u.Surname = surname
		u.City = city
		u.Age = age
		u.Content = content

		fmt.Println("Email:", email)

		user_list = append(user_list, u)

	}

	file.Close()

	t, err := template.ParseFiles("public/templates/bulk_email.html")
	log.Println("Parsing txt template")
	if err != nil {
		log.Println("Could not parse email template", err)
		return
	}

	dialer := setupDialer()

	for _, user := range user_list {

		var bodyBuffer bytes.Buffer

		if err := t.Execute(&bodyBuffer, gin.H{
			"title":   "OACS SERVICE",
			"body":    user.Content,
			"name":    user.Name,
			"surname": user.Surname,
			"city":    user.City,
			"age":     user.Age,
		}); err != nil {
			log.Println("Could not execute template", err)
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", "OACS@gmail.com")
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Hello!")
		m.SetBody("text/html", bodyBuffer.String())

		// Send the email
		if err := dialer.DialAndSend(m); err != nil {
			log.Printf("Could not send email to %s: %v", user.Email, err)
			// Decide how to handle the error - retry, continue, etc.
		}

	}

	c.String(200, "Success!")

}

func setupDialer() *gomail.Dialer {

	var host string
	var userName string
	var password string
	var port string

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Could not read the .env file!", err)

	}
	host = os.Getenv("HOST_NAME")
	userName = os.Getenv("EMAIL_NAME")
	password = os.Getenv("EMAIL_PASS")
	port = os.Getenv("EMAIL_PORT")
	iPort, _ := strconv.Atoi(port)

	return gomail.NewDialer(host, iPort, userName, password) //provider port 857
}

func SendFromDB(c *gin.Context) {

	log.Println("Got to SendFromDB!")

	var user_list []postgres.User_info

	sql := `SELECT surname, name, city, email, age FROM user_info`

	rows, err := postgres.GetDB().Query(sql)
	if err != nil {
		log.Println("Bad query at SendFromDB", err)
	}

	defer rows.Close()

	for rows.Next() {
		var user postgres.User_info

		if err = rows.Scan(&user.Surname, &user.Name, &user.City, &user.Email, &user.Age); err != nil {
			log.Println(err)
		}

		user_list = append(user_list, user)

	}

	t, err := template.ParseFiles("public/templates/bulk_email.html")
	log.Println("Parsing txt template")
	if err != nil {
		log.Println("Could not parse email template", err)
		return
	}

	log.Println("Code from DB processed!")

	dialer := setupDialer()

	for _, user := range user_list {

		var bodyBuffer bytes.Buffer

		if err := t.Execute(&bodyBuffer, gin.H{
			"title":   "OACS SERVICE",
			"body":    user.Content,
			"name":    user.Name,
			"surname": user.Surname,
			"city":    user.City,
			"age":     user.Age,
		}); err != nil {
			log.Println("Could not execute template", err)
			return
		}

		m := gomail.NewMessage()
		m.SetHeader("From", "OACS@gmail.com")
		m.SetHeader("To", user.Email)
		m.SetHeader("Subject", "Hello!")
		m.SetBody("text/html", bodyBuffer.String())

		// Send the email
		if err := dialer.DialAndSend(m); err != nil {
			log.Printf("Could not send email to %s: %v", user.Email, err)
			// Decide how to handle the error - retry, continue, etc.
		}

	}

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})

}
