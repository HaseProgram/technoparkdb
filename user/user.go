package user

import (
	"encoding/json"
	"database/sql"
	"github.com/go-ozzo/ozzo-routing"
	"technoparkdb/common"
)

type UserStruct struct {
	About string `json:"about"`
	Email string `json:"email"`
	Fullname string `json:"fullname"`
	Nickname string `json:"nickname"`
}

const insertStatement = "INSERT INTO users (about, email, fullname, nickname) VALUES ($1,$2,$3,$4)"
const selectStatement = "SELECT about, email, fullname, nickname FROM users WHERE email=$1 OR nickname=$2"
const selectStatementNickname = "SELECT about, email, fullname, nickname FROM users WHERE nickname=$1"
const selectStatementNicknameId = "SELECT id, nickname FROM users WHERE nickname=$1"

func getPost(c *routing.Context) UserStruct {
	var POST UserStruct
	c.Request.ParseForm();
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func GetUserId(nickname string, db *sql.DB) (int, string){
	row := db.QueryRow(selectStatementNicknameId, nickname)
	var id int
	err := row.Scan(&id, &nickname)
	switch err {
	case sql.ErrNoRows:
		return -1, ""
	case nil:
		return id, nickname
	default:
		panic(err)
	}
}

func Create(c *routing.Context, db *sql.DB) (string, int) {
	POST := getPost(c)
	defer c.Request.Body.Close()

	about := POST.About
	email := POST.Email
	fullname := POST.Fullname
	nickname := c.Param("nickname")

	row := db.QueryRow(insertStatement, about, email, fullname, nickname)
	err := row.Scan()
	if err != nil && err != sql.ErrNoRows {
		rows, selerr := db.Query("SELECT about, email, fullname, nickname FROM users WHERE email='" + email + "' OR nickname='" + nickname + "'")
		common.Check(selerr)

		var res []UserStruct

		for rows.Next() {
			var tus UserStruct
			err = rows.Scan(&tus.About, &tus.Email, &tus.Fullname, &tus.Nickname)
			common.Check(err)
			res = append(res, tus)
		}
		content, _ := json.Marshal(res)
		return string(content), 409
	}

	res := &UserStruct{
		About: about,
		Email: email,
		Fullname: fullname,
		Nickname: nickname,
	}

	content, _ := json.Marshal(res)
	return string(content), 201
}

func getProfile(nickname string, db *sql.DB) (string, int) {
	var res UserStruct
	row := db.QueryRow(selectStatementNickname, nickname)
	err := row.Scan(&res.About, &res.Email, &res.Fullname, &res.Nickname)
	switch err {
	case sql.ErrNoRows:
		var res common.ErrStruct
		res.Message = "User not found!"
		content, _ := json.Marshal(res)
		return string(content), 404
	case nil:
		content, _ := json.Marshal(res)
		return string(content), 200
	default:
		panic(err)
	}
}

func Profile(c *routing.Context, db *sql.DB) (string, int) {
	nickname := c.Param("nickname")
	return getProfile(nickname, db)
}

func Update(c *routing.Context, db *sql.DB) (string, int) {
	updateStatement := "UPDATE users SET"

	POST := getPost(c)
	defer c.Request.Body.Close()

	UPD := false

	about := POST.About
	if len(about) > 0 {
		updateStatement += " about='" + about
		UPD = true
	}
	email := POST.Email
	if len(email) > 0 {
		if UPD {
			updateStatement += "',"
		}
		updateStatement += " email='" + email
		UPD = true
	}
	fullname := POST.Fullname
	if len(fullname) > 0 {
		if UPD {
			updateStatement += "',"
		}
		updateStatement += " fullname='" + fullname
		UPD = true
	}

	nickname := c.Param("nickname")

	if UPD {
		updateStatement += "' WHERE nickname='" + nickname + "' RETURNING about, email, fullname, nickname"
		var resOk UserStruct
		err := db.QueryRow(updateStatement).Scan(&resOk.About, &resOk.Email, &resOk.Fullname, &resOk.Nickname)
		switch err {
		case sql.ErrNoRows:
			var resErr common.ErrStruct
			resErr.Message = "User not found!"
			content, _ := json.Marshal(resErr)
			return string(content), 404
		case nil:
			content, _ := json.Marshal(resOk)
			return string(content), 200
		default:
			var resErr common.ErrStruct
			resErr.Message = "Conflict while updating information!"
			content, _ := json.Marshal(resErr)
			return string(content), 409
		}
	}

	return getProfile(nickname, db)
}