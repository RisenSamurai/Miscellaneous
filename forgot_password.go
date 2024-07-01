package handlers

import (
	"Oaks/pkg/db"
	"Oaks/pkg/lib"
	"Oaks/pkg/session"
	"bytes"
	"gopkg.in/gomail.v2"
	"html/template"
	"log"
	"net/http"
	"time"
)

func ForgotPassword(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			log.Println("can't parse form!")
		}
		sql := `SELECT UserEmail FROM email_tokens WHERE UserEmail = ?`
		stmt, _ := db.GetDB().Prepare(sql)
		defer stmt.Close()
		var email string
		formEmail := r.FormValue("email")

		log.Println("formEmail = ", formEmail)

		err = stmt.QueryRow(formEmail).Scan(&email)

		if formEmail == email {
			log.Println("You can change your password only once an hour!")
			http.Redirect(w, r, "/", 303)
		} else if formEmail != email {
			log.Println("Got into the SendMail")

			var (
				From    string
				To      string
				Subject string
				Body    string
				Token   string
				Email   string
			)

			t, err := template.ParseFiles("web/templates/mail_template.html")
			log.Println("Parsing txt template")
			if err != nil {
				log.Println("Could not parse email template", err)
				return
			}

			Token, _ = lib.GenerateToken()

			From = "oacs.support@gmail.com"
			To = r.FormValue("email")
			Subject = "Password Reset"
			Body = "To reset your password, please click the following link:" +
				"f you did not request a password reset, please ignore this email." +
				"https://example.com/forgot/reset-password?token=" + Token + "\"</a>"

			sql = `SELECT Email FROM union_member WHERE Email = ?`

			stmt, err = db.GetDB().Prepare(sql)
			if err != nil {
				log.Println("No user with this email!", err)
			}
			defer stmt.Close()

			err = stmt.QueryRow(To).Scan(&Email)

			if Email == To {

				Sess.UserLogin = Email

				sql = `INSERT INTO email_tokens (Token, UserEmail) VALUES (?, ?)`

				stmt, err = db.GetDB().Prepare(sql)
				if err != nil {
					log.Println("Bad query at Token isnert!", err)
					return
				}
				defer stmt.Close()

				_, err = stmt.Exec(&Token, &To)

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

				log.Println("Sending....")

				d := gomail.NewDialer("NEO-HP2021", 25, "", "") //provider port 857

				// Send the email
				if err := d.DialAndSend(m); err != nil {
					log.Println("Could not send the email", err)
					return
				}

				Sess.Message = "Message was successfully sent!"

				http.Redirect(w, r, "/sign-in", 303)
			}

		}
	}

}

func CheckToken(w http.ResponseWriter, r *http.Request) {

	token := r.URL.Query().Get("token")
	currentTime := time.Now().UTC()
	var dbToken string
	var tokenTimeStamp string

	sql := `SELECT Token, CreatedTime FROM email_tokens WHERE Token = ?`
	stmt, err := db.GetDB().Prepare(sql)
	defer stmt.Close()
	err = stmt.QueryRow(token).Scan(&dbToken, &tokenTimeStamp)
	if err == nil {
		tokenTime, err := time.Parse("2006-01-02 15:04:05", tokenTimeStamp)
		if err != nil {
			log.Println("Can't parse the timestamp!")
		}

		expirationTime := tokenTime.Add(60 * time.Minute)

		if token == dbToken {
			if currentTime.After(expirationTime) {
				log.Println("Token is expired!")
			} else {
				Sess.ForgotPassword = true
				http.Redirect(w, r, "/forgot", 303)
			}

		}
	} else {
		log.Println("You can change you password only once a day!")
	}

}

func ChangePassword(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {

		err := r.ParseForm()
		if err != nil {
			log.Println("Can't parse form!")
		}

		formPassword := r.FormValue("password")
		formRepeat := r.FormValue("passwordrep")
		var ID int64

		if formPassword == formRepeat {

			newPassword, _ := lib.HashPassword(formPassword)

			sql := `SELECT UserID from users WHERE email = ?`

			stmt, err := db.GetDB().Prepare(sql)
			if err != nil {
				log.Println("Bad query at Change password!", err)
			}

			defer stmt.Close()

			err = stmt.QueryRow(Sess.UserLogin).Scan(&ID)

			sql = `UPDATE passwords  SET UserHash = ? WHERE UserID = ?`
			stmt, err = db.GetDB().Prepare(sql)
			if err != nil {
				log.Println("Can't update user hash!")
			}
			defer stmt.Close()

			_, err = stmt.Exec(newPassword, ID)

			http.Redirect(w, r, "/sign-in", 303)

		} else {
			log.Println("Passwords are not matching!")
		}

	}

}

func ForgotDisplay(w http.ResponseWriter, r *http.Request) {

	session.CreateSession(w, r)

	data := lib.PageData{
		HeaderData: lib.Header{
			Title: "OACS | Reset Password",
		},
		BodyData: Sess,
		FooterData: lib.Footer{
			CopyrightYear: 2023,
		},
	}

	lib.RenderPage(w, "forgot_password.html", data)
}
