package admin

import (
	"Oaks/pkg/db"
	"Oaks/pkg/handlers"
	"Oaks/pkg/lib"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type PinRec struct {
	Pin      string
	MemberID string
}
type MemberDetails struct {
	MemberID string
}

func MemberCodeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("THis is MemberCodeHandler")

	if r.Method == http.MethodPost {

		choose := r.FormValue("select")

		log.Println("Received: ", choose)

		if choose == "2" {

			rowsAmount := r.FormValue("amount")
			rowsAmountInt, _ := strconv.Atoi(rowsAmount)
			year, _ := strconv.Atoi(r.FormValue("year"))
			pinLength, _ := strconv.Atoi(r.FormValue("length"))
			//Filling the database with pin codes
			log.Println("Generating pins...")
			_, err := makePinTable(rowsAmountInt, year, pinLength)
			if err != nil {
				log.Println("Can't generate pins!", err)
				return
			}

			http.Redirect(w, r, "/user/pin-codes", 303)

		}
		if choose == "one" {
			//Inserting one row into NULL
			memberNumber := r.FormValue("memberNumber")
			returnedPin, err := InsertOneRowToDb(memberNumber)

			log.Println("Received memberNumber:", memberNumber)
			if err != nil {
				log.Println("Error from InsertOneRowToDb:", err)
				return
			}
			//url := fmt.Sprintf("/Service/pin=%s&message=%s", url.QueryEscape(returnedPin), url.QueryEscape("Data was inserted and successfully returned"))
			handlers.Sess.Message = returnedPin
			http.Redirect(w, r, "/kakunin/pincode", http.StatusSeeOther)

		}
		if choose == "file" {
			//Loading a lot of Member Numbers in NULL positions
		}

		log.Println("Pins were generated!")

	}

}

// InsertOneRowToDb Taking the pin code by memberID and putting memberID into NULL
func InsertOneRowToDb(memberID string) (string, error) {

	var memberCheck string
	err := db.GetDB().QueryRow("SELECT UserID FROM pins WHERE UserID = ?", memberID).Scan(&memberCheck)
	if err != nil {
		log.Println("SQL Query error")
	}
	if memberCheck == memberID {
		log.Println("User with this memberID is already exists!")
		return "User with this memberID is already exists!", err
	} else {

		_, err = db.GetDB().Exec("UPDATE pins SET UserID = ? WHERE UserID IS NULL LIMIT 1", memberID)
		if err != nil {
			return "bad query", err
		}
		var pin string
		err = db.GetDB().QueryRow("SELECT pin FROM pins WHERE UserID = ?", memberID).Scan(&pin)
		if err != nil {
			return "bad query:", err
		}

		log.Println("Data was successfully inserted!")
		return pin, nil
	}

}

func makePinTable(amount int, year int, pinLength int) (message string, err error) {

	for i := 0; i < amount; i++ {
		number, err := generatePin(pinLength)
		if err != nil {
			return "Could not to generate pin!", err
		}
		_, err = db.GetDB().Exec("INSERT INTO pins (Pin, RecordDate) VALUES (?, ?)", number, year)
		if err != nil {
			return "Failed to insert pin", err
		}
	}

	rows, err := db.GetDB().Query("SELECT UnionMemberNumber  FROM union_member")
	if err != nil {
		return "Failed to fetch union members", err
	}
	defer rows.Close()

	rows2, err := db.GetDB().Query("SELECT Pin FROM pins")
	if err != nil {
		return "Failed to fetch pins", err
	}
	defer rows2.Close()

	if rows.Next() {
		res := MemberDetails{}
		err = rows.Scan(&res.MemberID)
		if err != nil {
			return "Failed to scan MemberID", err
		}
		// read pin
		if rows2.Next() {
			res2 := &PinRec{}
			err = rows2.Scan(&res2.Pin)
			if err != nil {
				return "Failed to scan pin", err
			}
			res2.MemberID = res.MemberID
			// update pintable
			upd, err := db.GetDB().Prepare("UPDATE pins set UnionMemberNumber = ? where Pin = ? ")
			if err != nil {
				return "Failed to prepare update query", err
			}
			_, err = upd.Exec(res.MemberID, res2.Pin)
			if err != nil {
				return "Failed to update pin", err
			}
		}
	}

	return "Successfully created pin table and populated data", nil
}

func generatePin(pinLength int) (int, error) {
	// Validate pinLength
	if pinLength <= 0 {
		return 0, errors.New("invalid pin length")
	}

	// Define the lower and upper bounds for the pin
	lowerBound := int(math.Pow10(pinLength - 1))
	upperBound := int(math.Pow10(pinLength)) - 1

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate the pin
	pin := rand.Intn(upperBound-lowerBound+1) + lowerBound

	log.Println("Generated PIN:", pin)
	return pin, nil
}

func InsertDataToDb(pin string, memberID string) error {

	_, err := db.GetDB().Exec("UPDATE pins SET UserID = ? WHERE Pin = ? AND UserID IS NULL LIMIT 1", memberID, pin)
	if err != nil {
		return fmt.Errorf("bad query: %w", err)
	}

	log.Println("Data was successfully inserted!")
	return nil
}

// SelectTheFirstNull Getting the first null row
func SelectTheFirstNull() (message string, err error) {

	var pinRec int

	err = db.GetDB().QueryRow("SELECT * FROM pins WHERE UserID IS NULL LIMIT 1").Scan(&pinRec)
	if err != nil {
		if err == sql.ErrNoRows {
			return "No NULL rows found", nil
		}
		return "Failed to query for NULL rows", err
	}

	return fmt.Sprintf("Found a NULL row with pinRec: %d", pinRec), nil
}

func PincodesDisplay(w http.ResponseWriter, r *http.Request) {

	data := lib.PageData{
		HeaderData: lib.Header{Title: "OACS | Generate pin"},
		BodyData:   handlers.Sess,
	}

	lib.RenderPage(w, "pin_code_selection.html", data)

}
