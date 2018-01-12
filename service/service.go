package service

import (
	"github.com/go-ozzo/ozzo-routing"
	"github.com/HaseProgram/technoparkdb/database"
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
	var sttms []string

	deleteStatement := "DELETE FROM forums"
	sttms = append(sttms, deleteStatement)
	deleteStatement = "DELETE FROM threads"
	sttms = append(sttms, deleteStatement)
	deleteStatement = "DELETE FROM posts"
	sttms = append(sttms, deleteStatement)
	deleteStatement = "DELETE FROM forum_users"
	sttms = append(sttms, deleteStatement)
	deleteStatement = "DELETE FROM thread_votes"
	sttms = append(sttms, deleteStatement)
	deleteStatement = "DELETE FROM users"
	sttms = append(sttms, deleteStatement)

	for _, s := range sttms {
		err := db.QueryRow(s).Scan()
		if err != pgx.ErrNoRows {
			return "", 409
		}
	}

	return "", 200
}