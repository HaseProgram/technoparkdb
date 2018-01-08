package service

import (
	"github.com/go-ozzo/ozzo-routing"
	"technoparkdb/database"
	"encoding/json"
	"github.com/jackc/pgx"
)

type CountStruct struct {
	Forum int `json:"forum"`
	Post int `json:"post"`
	Thread int `json:"thread"`
	User int `json:"user"`
}

func Status(c *routing.Context) (string, int) {
	db := database.DB
	selectStatement := `
		SELECT uc.*, fc.*, tc.*, pc.* FROM
		(SELECT count(*) FROM users) AS uc,
		(SELECT count(*) FROM forums) AS fc,
		(SELECT count(*) FROM threads) AS tc,
		(SELECT count(*) FROM posts) AS pc
	`
	row := db.QueryRow(selectStatement)
	var counter CountStruct
	err := row.Scan(&counter.User, &counter.Forum, &counter.Thread, &counter.Post)
	switch err {
	case nil:
		content, _ := json.Marshal(counter)
		return string(content), 200
	default:
		panic(err)
	}
}

func Clear(c *routing.Context) (string, int) {
	db := database.DB
	deleteStatement := "DELETE FROM users"
	err := db.QueryRow(deleteStatement).Scan()
	if err == pgx.ErrNoRows {
		return "", 200
	}
	return "", 409
}