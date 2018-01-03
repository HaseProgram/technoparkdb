package thread

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
	"encoding/json"
	"technoparkdb/common"
	"technoparkdb/user"
	"time"
)

type ThreadStruct struct {
	Id int `json:"id"`
	Slug string `json:"slug"`
	Author string `json:"author"`
	Title string `json:"title"`
	ForumSlug string `json:"forum"`
	Message string `json:"message"`
	Created time.Time `json:"created"`
}

const insertStatement = "INSERT INTO threads (author_id, author_name, forum_id, forum_slug, title, created, message, slug) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
const selectStatementSlug = "SELECT id, slug, created, message, title FROM threads WHERE slug=$1"
const selectStatementForumSlugId = "SELECT id, slug FROM forums WHERE slug=$1"

func getPost(c *routing.Context) ThreadStruct {
	var POST ThreadStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func GetForumSlugId(slug string, db *sql.DB) (int, string){
	row := db.QueryRow(selectStatementForumSlugId, slug)
	var id int
	err := row.Scan(&id, &slug)
	switch err {
	case sql.ErrNoRows:
		return -1, ""
	case nil:
		return id, slug
	default:
		panic(err)
	}
}

// Сделать сначала select, потом insert

func CreateThread(c *routing.Context, db *sql.DB) (string, int) {
	forumSlug := c.Param("slug")
	forumId := -1
	POST := getPost(c)
	defer c.Request.Body.Close()

	author := POST.Author
	authorId := -1
	created := POST.Created
	title := POST.Title
	message := POST.Message
	slug := POST.Slug

	authorId, author = user.GetUserId(author, db)
	if authorId >= 0 {
		forumId, forumSlug = GetForumSlugId(forumSlug, db)
		if forumId >= 0 {
			var res ThreadStruct
			res.Author = author
			res.ForumSlug = forumSlug

			//костыль, чтобы вставлять в базу nil, а не пустую строку
			err := sql.ErrNoRows
			if len(slug) > 0 {
				row := db.QueryRow(selectStatementSlug, slug)
				err = row.Scan(&res.Id, &res.Slug, &res.Created, &res.Message, &res.Title)
			}

			switch err {
			case sql.ErrNoRows:
				var row *sql.Row
				// костыль продолжается
				if len(slug) > 0 {
					row = db.QueryRow(insertStatement, authorId, author, forumId, forumSlug, title, created, message, slug)
				} else {
					row = db.QueryRow(insertStatement, authorId, author, forumId, forumSlug, title, created, message, nil)
				}
				err := row.Scan(&res.Id)
				common.Check(err)
				res.Created = created
				res.Message = message
				res.Slug = slug
				res.Title = title
				content, _ := json.Marshal(res)
				return string(content), 201
			case nil:
				content, _ := json.Marshal(res)
				return string(content), 409
			default:
				panic(err)
			}
		}
	}
	var res common.ErrStruct
	res.Message = "Can not create thread. Thread author not found"
	content, _ := json.Marshal(res)
	return string(content), 404

}