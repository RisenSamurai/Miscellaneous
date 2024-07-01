package postgres

type Users struct {
	id       int64  `json:"id"`
	Email    string `json:"Email"`
	Password string `json:"Password"`
	Role     string `json:"Role"`
}

type crypto struct {
	id          int64  `db:"id"`
	user_id     int64  `db:"user_id"`
	password    string `db:"password"`
	create_date string `db:"create_date"`
}

type qr struct {
	id            int64  `db:"id"`
	user_id       int64  `db:"user_id"`
	qr_path       string `db:"qr_path"`
	date_creation string `db:"date_creation"`
}

type token struct {
	id            int64  `db:"id"`
	user_id       int64  `db:"user_id"`
	token         string `db:"token"`
	creation_date string `db:"creation_date"`
}

type User_info struct {
	Id             int64  `db:"id"`
	Surname        string `db:"surname"`
	Name           string `db:"name"`
	City           string `db:"city"`
	Email          string `db:"email"`
	Age            string `db:"Age"`
	BirthDate      string `db:"birthdate"`
	MarriageStatus string
	Content        string
}
