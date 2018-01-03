package user

import (
	"encoding/json"
	"database/sql"
	"github.com/go-ozzo/ozzo-routing"
	"time"
	"fmt"
)

type UserStruct struct {
	About string `json:"about"`
	Email string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
}

var POST UserStruct

const insertStatement = "INSERT INTO users (about, email, fullname, nickname) VALUES ($1,$2,$3,$4)"
const selectStatement = "SELECT about, email, fullname, nickname FROM users WHERE email=$1 OR nickname=$2"

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func getPost(c *routing.Context) {
	c.Request.ParseForm();
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	check(err)
}

func Create(c *routing.Context, db *sql.DB) (string, int) {
	getPost(c)
	defer c.Request.Body.Close()

	about := POST.About
	email := POST.Email
	fullname := POST.Fullname
	nickname := c.Param("nickname")

	code := 201
	row := db.QueryRow(insertStatement, about, email, fullname, nickname)
	err := row.Scan()
	if err != nil && err != sql.ErrNoRows {
		code = 409
		rows, err := db.Query("SELECT about, email, fullname, nickname FROM users WHERE email=" + email + " OR nickname=" + nickname)
		check(err)

		for rows.Next() {
			var uid int
			var username string
			var department string
			var created time.Time
			err = rows.Scan(&uid, &username, &department, &created)
			check(err)
		}
	}

	res := &UserStruct{
		About: about,
		Email: email,
		Fullname: fullname,
		Nickname: nickname,
	}

	content, _ := json.Marshal(res)
	return string(content), code
}