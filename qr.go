package server

import (
	"bytes"
	"github.com/skip2/go-qrcode"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"net/http"
	"oacs/server/lib"
	"oacs/server/old/db"
)

func SendQR(w http.ResponseWriter, r *http.Request) {

	log.Println("Got into the SendMail")

	sql := `SELECT  u.email, q.QrName, q.UserID
            FROM users AS u
            LEFT JOIN qr AS q
            ON u.UserID = q.UserID
            WHERE q.SentFlag IS NULL AND q.UserID IS NOT NULL`

	stmt, err := db.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query at SendQR", err)
	}

	defer stmt.Close()

	rows, err := stmt.Query()

	for rows.Next() {

		var (
			From    string
			To      string
			Subject string
			Body    string
			Email   string
			QrName  string
			UserID  int
		)

		if err = rows.Scan(&Email, &QrName, &UserID); err != nil {
			log.Println("Cant scan rows!", err)
		}

		sql := `UPDATE qr SET SentFlag = 1 WHERE UserID = ?`
		stmt, err := db.GetDB().Prepare(sql)

		_, err = stmt.Exec(UserID)
		if err != nil {
			log.Println("Can't execute SentFlag", err)
		}

		t, err := template.ParseFiles("web/templates/mail_qr_template.html")
		log.Println("Parsing txt template")
		if err != nil {
			log.Println("Could not parse email template", err)
			return
		}

		log.Println("Parsing file template")

		From = "oacs.support@gmail.com"
		To = Email
		Subject = "OACS system sign-up"
		Body = "Dear user please sign up in our system by scanning a QR code below."

		data := struct {
			RecipientName string
			Body          string
			Subject       string
		}{
			RecipientName: To,
			Body:          Body,
			Subject:       Subject,
		}

		var bodyBuffer bytes.Buffer

		if err := t.Execute(&bodyBuffer, data); err != nil {
			log.Println("Could not execute template", err)
			return
		}
		log.Println("Assigning fields")
		m := gomail.NewMessage()
		m.SetHeader("From", From)
		m.SetHeader("To", To)
		m.SetHeader("Subject", Subject)
		m.SetBody("text/html", bodyBuffer.String())
		m.Embed("QR/"+QrName, gomail.SetHeader(map[string][]string{
			"Content-ID": {"myimagecid"},
		}))

		log.Println("Sending....")

		d := gomail.NewDialer("NEO-HP2021", 25, "", "") //provider port 857

		// Send the email
		if err := d.DialAndSend(m); err != nil {
			log.Println("Could not send the email", err)
			return
		}

	}

	defer rows.Close()

}

func GenerateQR(w http.ResponseWriter, r *http.Request) {

	var sql string
	var err error

	sql = `SELECT p.id, p.Password, u.Email
		   FROM crypto AS p
           LEFT JOIN users AS u 
    	   ON p.UserID = u.UserID
           WHERE p.UserID is NOT NULL
		   `

	stmt, err := db.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query at GenerateQR select email", err)
	}

	defer stmt.Close()

	rows, err := stmt.Query()

	sql = `INSERT INTO qr (UserID, QrName) VALUES (?, ?)`

	for rows.Next() {

		var user lib.User

		if err = rows.Scan(&user.UserID, &user.UserPin, &user.UserEmail); err != nil {
			log.Println("Unable to read rows", err)
		}

		err = qrcode.WriteFile("http://192.168.1.17:8080/client/matching?pin="+user.UserPin,
			qrcode.Medium, 256, "QR/"+user.UserEmail+".qr.png")
		if err != nil {
			log.Println("Could not encode the qr code!", err)
		}

		stmt, err := db.GetDB().Prepare(sql)
		if err != nil {
			log.Println("Bad QR data insertion", err)
		}

		_, err = stmt.Exec(user.UserID, user.UserEmail+".qr.png")

	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

}
