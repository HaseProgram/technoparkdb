package forum

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
	"encoding/json"
	"technoparkdb/common"
)

type ForumStruct struct {
	User string `json:"user"`
	Title string `json:"title"`
	Slug string `json:"slug"`
	Posts int `json:"posts_count"`
	Threads int `json:"threads_count"`
}

const insertStatement = "INSERT INTO forums (owner_id, owner_nickname, title, slug) VALUES ($1,$2,$3,$4)"
const selectStatementNickname = "SELECT id, nickname FROM users WHERE nickname=$1"
const selectStatementSlug = "SELECT slug, title FROM forums WHERE slug=$1"
const selectStatementSlugAll = "SELECT slug, title, owner_nickname, posts_count, threads_count FROM forums WHERE slug=$1"

func getPost(c *routing.Context) ForumStruct {
	var POST ForumStruct
	c.Request.ParseForm();
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func getUser(nickname string, db *sql.DB) (int, string){
	row := db.QueryRow(selectStatementNickname, nickname)
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

	slug := POST.Slug
	title := POST.Title
	user := POST.User
	user_id := -1

	user_id, user = getUser(user, db)
	if user_id >= 0 {
		var res ForumStruct
		row := db.QueryRow(insertStatement, user_id, user, title, slug)
		err := row.Scan()
		if err != nil && err != sql.ErrNoRows {
			row := db.QueryRow(selectStatementSlug, slug)
			err := row.Scan(&res.Slug, &res.Title)
			res.User = user
			switch err {
			case nil:
				content, _ := json.Marshal(res)
				return string(content), 409
			default:
				panic(err)
			}
		}
		res.User = user
		res.Slug = slug
		res.Title = title
		res.Threads = 0;
		res.Posts = 0;
		content, _ := json.Marshal(res)
		return string(content), 201
	}
	var res common.ErrStruct
	res.Message = "Can not create forum. User-creator not found"
	content, _ := json.Marshal(res)
	return string(content), 404
}

func Details(c *routing.Context, db *sql.DB) (string, int) {
	slug := c.Param("slug")
	var res ForumStruct

	row := db.QueryRow(selectStatementSlugAll, slug)
	err := row.Scan(&res.Slug, &res.Title, &res.User, &res.Posts, &res.Threads)
	switch err {
	case nil:
		content, _ := json.Marshal(res)
		return string(content), 200
	case sql.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Forum not found!"
		content, _ := json.Marshal(res)
		return string(content), 404
	default:
		panic(err)
	}
}