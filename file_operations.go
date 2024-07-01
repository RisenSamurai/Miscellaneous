package admin

import (
	"Oaks/pkg/db"
	"Oaks/pkg/handlers"
	"Oaks/pkg/lib"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func WriteFromDbToCSV(w http.ResponseWriter, r *http.Request) {

	var sql string
	var err error

	file, err := os.Create("service/UserData.csv")
	if err != nil {
		log.Println("Can't create the file", err)
	}
	defer file.Close()

	sql = `SELECT u.UserID,u.email, p.UserHash, u.UserRole, m.HealthInsuranceNumber
		   FROM users AS u 
		   LEFT JOIN passwords AS p
		   ON u.UserID = p.UserID
		   LEFT JOIN union_member AS m 
		   ON  u.UserID = m.UserID`

	stmt, err := db.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query at db to csv!", err)
	}
	rows, err := stmt.Query()

	for rows.Next() {

		var U lib.UserToCSV

		if err = rows.Scan(&U.UserID, &U.UserEmail, &U.UserHash, &U.UserRole, &U.HealthInsuranceNumber); err != nil {
			log.Println("Unable to read rows", err)
		}

		file.WriteString(fmt.Sprintf("UserID %d \n UserEmail %s \n UserHash %s \n "+
			"UserRole %s \n HealthInsuranceNumber %s \n", U.UserID, U.UserEmail, U.UserHash, U.UserRole,
			U.HealthInsuranceNumber))

	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	defer rows.Close()

}

func WriteFromCSVToDb(w http.ResponseWriter, r *http.Request) {

	file, err := os.Open("service/UserData.csv")
	if err != nil {
		log.Println("Can't read the file!")
	}
	defer file.Close()

	var batch []lib.UserToCSV

	batchSize := 1000
	var dataRecord lib.UserToCSV

	batch = make([]lib.UserToCSV, 0, batchSize)

	reader := csv.NewReader(file)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		dataRecord = lib.UserToCSV{
			UserID:                record[0],
			UserRole:              record[1],
			UserEmail:             record[2],
			UserHash:              record[3],
			HealthInsuranceNumber: record[4],
		}

		batch = append(batch, dataRecord)

		if len(batch) >= batchSize {
			batch = make([]lib.UserToCSV, 0, batchSize)
		}

		if len(batch) > 0 {

			dataRecord = lib.UserToCSV{
				UserID:                record[0],
				UserRole:              record[1],
				UserEmail:             record[2],
				UserHash:              record[3],
				HealthInsuranceNumber: record[4],
			}
		}

		if err != nil {
			log.Println("can't read the file!", err)
		}

	}
}

func FilesUpload(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Println("Can't parse form", err)
		return
	}

	files := r.MultipartForm.File["file"]

	for _, file := range files {
		src, err := file.Open()
		if err != nil {
			log.Println("Can't open the file!", err)
		}
		defer src.Close()

		//fileSize := file.Size

		err = os.Mkdir("uploads/"+handlers.Sess.UserLogin, 0750)
		if err != nil && !os.IsExist(err) {
			log.Println("Can't create the folder!", err)
		}

		dst, err := os.Create("uploads/" + handlers.Sess.UserLogin + "/" + file.Filename)
		if err != nil {
			log.Println("Cant write file to a folder!", err)
			return
		}

		defer dst.Close()

		fileBytes, err := io.ReadAll(src)
		if err != nil {
			log.Println("Can't write file down!", err)
		}

		//UploadFilesData(file.Filename, fileSize)

		dst.Write(fileBytes)

		handlers.Sess.Img_path = "uploads/" + handlers.Sess.UserLogin + "/" + file.Filename
		w.WriteHeader(200)

	}

}

func UploadFilesData(fileName string, fileSize int64) {

	sql := `INSERT INTO file (FileName, FileSize)
			VALUES (?, ?)`

	stmt, err := db.GetDB().Prepare(sql)
	if err != nil {
		log.Println("Bad query at UploadFilesData", err)
		return
	}

	defer stmt.Close()

	_, err = stmt.Exec(fileName, fileSize)
	if err != nil {
		log.Println("Unable to execute query!", err)
	}

}

func FilesDisplay(w http.ResponseWriter, r *http.Request) {

	data := lib.PageData{
		HeaderData: lib.Header{
			Title:   "OACS | File Upload",
			IsAdmin: true,
		},
		BodyData: handlers.Sess,
		FooterData: lib.Footer{
			CopyrightYear: 2023,
		},
	}

	lib.RenderPage(w, "load_files.html", data)
}
