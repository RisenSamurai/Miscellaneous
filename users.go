package server

import (
	"encoding/csv"
	"log"
	"net/http"
	"oacs/server/lib"
	"oacs/server/postgres"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetUsers(c *gin.Context) {

	log.Println("Got into GetUsers!")
	var users []postgres.User_info

	sql := `SELECT id, surname, name, birthdate FROM user_info LIMIT 10`

	rows, err := postgres.GetDB().Query(sql)
	if err != nil {
		log.Println("Bad query at Get Users", err)
	}

	defer rows.Close()

	for rows.Next() {

		var u postgres.User_info

		if err := rows.Scan(&u.Id, &u.Surname, &u.Name, &u.BirthDate); err != nil {
			log.Println("Error during scanning!", err)

		}

		users = append(users, u)

	}

	c.JSON(http.StatusOK, users)

}

func GetUser(c *gin.Context) {

	var u postgres.User_info

	userID := c.Param("userId")

	sql := "SELECT surname, name, city, birthdate, age, marriage FROM user_info WHERE id = $1"

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query at GetUser!", err)
		return
	}

	defer stmt.Close()

	err = stmt.QueryRow(userID).Scan(&u.Surname, &u.Name, &u.City, &u.BirthDate, &u.Age,
		&u.MarriageStatus)
	if err != nil {
		log.Println("Bad query!", err)
		return
	}

	c.JSON(http.StatusOK, u)

}

func createUserCSV(c *gin.Context) string {

	var u postgres.User_info
	id := c.Param("userId")

	sql := `SELECT id, surname, name, city, email, age, birthdate, marriage
			FROM user_info WHERE id = $1`

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query!", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(id).Scan(&u.Id, &u.Surname, &u.Name, &u.City, &u.Email, &u.Age,
		&u.BirthDate, &u.MarriageStatus)

	rows := [][]string{
		{"Surname", "Name", "City", "Email", "Age", "BirthDate", "Marriage"},
		{u.Surname, u.Name, u.City, u.Email, u.Age, u.BirthDate, u.MarriageStatus},
	}

	surname := strings.TrimSpace(u.Surname)
	name := strings.TrimSpace(u.Name)

	os.Mkdir(filepath.Join("public", "csv"), 0755)
	file, err := os.Create(filepath.Join("public", "csv", surname+name+".csv"))
	if err != nil {
		log.Print("Can't create a file!", err)
	}

	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, row := range rows {

		if err := writer.Write(row); err != nil {
			log.Println("Can't write a row to a file!", err)
		}
	}

	file_path := filepath.Join("public", "csv", surname+name+".csv")

	log.Println(file_path)

	return file_path

}

func GetUserCSV(c *gin.Context) {

	file_path := createUserCSV(c)

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", "attachment; filename=DowGreg.csv")
	c.Header("Content-Type", "text/csv")
	c.File(file_path)

}

func CreateUser(c *gin.Context) {

	var u postgres.Users
	var lastID int64
	err := c.BindJSON(&u)
	if err != nil {
		log.Println(err)
	}

	log.Println("email: ", u.Email)

	sql := `INSERT INTO users (email, role) VALUES ($1, $2) RETURNING id`

	stmt, err := postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad prepare at CreateUser", err)
		return
	}

	defer stmt.Close()

	err = stmt.QueryRow(u.Email, u.Role).Scan(&lastID)

	_, err = stmt.Exec(u.Email, u.Role)
	if err != nil {
		log.Println("Bad query! here", err)
	}

	sql = `INSERT INTO crypto (user_id, password) VALUES ($1, $2)`

	stmt, err = postgres.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad prepare!", err)
	}

	defer stmt.Close()

	hashedP, err := lib.HashPassword(u.Password)
	if err != nil {
		log.Println("Can't hash the password!", err)
	}
	if err != nil {

	}

	_, err = stmt.Exec(lastID, hashedP)
	if err != nil {
		log.Println("Bad query!", err)
	}

	c.JSON(200, gin.H{
		"message": "New user has been created!",
	})

}
