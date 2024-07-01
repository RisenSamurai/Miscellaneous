package routines

import (
	"Oaks/pkg/db"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"time"
)

func MakeBackUp() {

	log.Println("Task is running", time.Now().String())
	MakeMembersBackUp()
	MakePinsBackUp()
}

func BackUpScheduler() {
	c := cron.New(cron.WithSeconds())
	_, err := c.AddFunc("0 30 10 * * ?", MakeBackUp) // Runs every day at 12:00pm
	if err != nil {
		log.Fatal(err)
	}
	c.Start()

}

func MakeMembersBackUp() {

	// Query data from table
	rows, err := db.GetDB().Query("SELECT * FROM union_member")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	// Create a backup file
	file, err := os.Create("backup.sql")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Loop through rows and write to file
	for rows.Next() {
		var (
			memberID             string
			memberName           string
			memberNameFurigana   string
			memberGender         string
			memberBirthDay       string
			isCurrent            string
			unionNumber          string
			unionMemberNumber    string
			memberStatus         string
			dateOfJoining        string
			isMerried            string
			dateOfResignation    string
			survayStartDay       string
			survayCompletionDate string
			survayStatus         string
			workPhoneNumber      string
			email                string
			survayDependentName  string
			updateDate           string
			Address              string
			homePhoneNumber      string
		)
		if err := rows.Scan(&memberID, &memberName, &memberNameFurigana, &memberGender,
			&memberBirthDay, &isCurrent, &unionNumber, &unionMemberNumber, &memberStatus,
			&dateOfJoining, &isMerried, &dateOfResignation, &survayStartDay, &survayCompletionDate,
			&survayStatus, &workPhoneNumber, &homePhoneNumber, &email, &survayDependentName,
			&updateDate, &Address); err != nil {

		}
		file.WriteString(fmt.Sprintf("\nMemberID: %s\nMemberName: %s\nMemberNameFurigana: %s\nMemberGender: %s\n"+
			"MemberBirthDay:  %s\nisCurrent: %s\nunionNumber: %s\nUnionMemberNumber: %s\nMemberStatus: %s\n"+
			"DateOfJoining: %s\nisMerried: %s\nDateOfResignation: %s\nSurvayStartDate: %s\nsurvayStatus: %s\n"+
			"SurvayCompletionDate: %s\nworkPhoneNumber: %s\nhomePhoneNumber: %s\nEmail: %s\nSurvayDependentName: %s\n"+
			"UpdateDate: %s\nAddress: %s\n",
			memberID, memberName, memberNameFurigana, memberGender, memberBirthDay, isCurrent, unionNumber,
			unionMemberNumber, memberStatus, dateOfJoining, isMerried, dateOfResignation, survayStartDay,
			survayStatus, survayCompletionDate, workPhoneNumber, homePhoneNumber, email, survayDependentName,
			updateDate, Address,
		))
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
}

func MakePinsBackUp() {
	rows, err := db.GetDB().Query("SELECT * FROM pincodes")
	if err != nil {
		log.Println("Bad Query!", err)
	}
	defer rows.Close()

	file, err := os.Create("pincodesBackUp.sql")
	if err != nil {
		log.Println("Can't create a file!", err)
	}
	defer file.Close()

	for rows.Next() {
		var (
			Pin       string
			MemberID  string
			YearCycle string
		)

		if err := rows.Scan(&Pin, &MemberID, &YearCycle); err != nil {
			log.Println("Can't write data to the file!")
		}

		file.WriteString(fmt.Sprintf("\nPin: %s\nMemberID: %s\nYearCycle: %s\n", Pin, MemberID, YearCycle))
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

	}

}

func MakeDependentsBackUp(w http.ResponseWriter, r *http.Request) {

}
