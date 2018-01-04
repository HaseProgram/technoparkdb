package thread

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
	"encoding/json"
	"technoparkdb/common"
	"technoparkdb/user"
	"time"
	"strconv"
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
const selectStatementSlug = "SELECT id, slug, created, message, title, author_name, forum_slug FROM threads WHERE slug=$1"
const selectStatementSlugOrID = "SELECT id, slug, created, message, title, author_name, forum_slug FROM threads WHERE slug=$1 OR id=$2"
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

			//костыль, чтобы вставлять в базу nil, а не пустую строку. мб позже попробовать с дефолтным параметром
			err := sql.ErrNoRows
			if len(slug) > 0 {
				row := db.QueryRow(selectStatementSlug, slug)
				err = row.Scan(&res.Id, &res.Slug, &res.Created, &res.Message, &res.Title, &res.Author, &res.ForumSlug)
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
				res.Author = author
				res.ForumSlug = forumSlug
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

func getThread(slug string, ids int, db *sql.DB) (string, int) {
	var res ThreadStruct
	row := db.QueryRow(selectStatementSlugOrID, slug, ids)
	err := row.Scan(&res.Id, &res.Slug, &res.Created, &res.Message, &res.Title, &res.Author, &res.ForumSlug)
	switch err {
	case sql.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Thread not found!"
		content, _ := json.Marshal(res)
		return string(content), 404
	case nil:
		content, _ := json.Marshal(res)
		return string(content), 200
	default:
		panic(err)
	}
}

func Details(c *routing.Context, db *sql.DB) (string, int) {
	slug := c.Param("slugid")
	id, _ := strconv.Atoi(slug)
	return getThread(slug, id, db)
}

func Update(c *routing.Context, db *sql.DB) (string, int) {
	updateStatement := "UPDATE threads SET"

	POST := getPost(c)
	defer c.Request.Body.Close()

	UPD := false

	message := POST.Message
	if len(message) > 0 {
		updateStatement += " message='" + message
		UPD = true
	}
	title := POST.Title
	if len(title) > 0 {
		if UPD {
			updateStatement += "',"
		}
		updateStatement += " title='" + title
		UPD = true
	}

	slug := c.Param("slugid")
	var ids int
	ids, err := strconv.Atoi(slug)
	orstate := ""
	if err == nil {
		orstate = " id=" + slug
	} else {
		orstate = " slug='" + slug + "'"
	}
	if UPD {
		updateStatement += "' WHERE" + orstate + " RETURNING author_name, created, forum_slug, id, message, title, slug"
		var resOk ThreadStruct
		err := db.QueryRow(updateStatement).Scan(&resOk.Author, &resOk.Created, &resOk.ForumSlug, &resOk.Id, &resOk.Message, &resOk.Title, &resOk.Slug)
		switch err {
		case sql.ErrNoRows:
			var resErr common.ErrStruct
			resErr.Message = "Thread not found!"
			content, _ := json.Marshal(resErr)
			return string(content), 404
		case nil:
			content, _ := json.Marshal(resOk)
			return string(content), 200
		}
	}
	return getThread(slug, ids, db)
}

func GetThreads(c *routing.Context, db *sql.DB) (string, int) {
	forumSlug := c.Param("slug")
	forumId, forumSlug := GetForumSlugId(forumSlug, db)
	if forumId >= 0 {
		selectStatementThreads := "SELECT id, author_name, title, created, message, slug FROM threads WHERE forum_slug='" + forumSlug + "'"

		desc := c.Query("desc")
		since := c.Query("since")
		if len(since) > 0 {
			switch desc {
			case "true":
				selectStatementThreads += " AND created <= '" + since + "'"
			default:
				selectStatementThreads += " AND created >= '" + since + "'"
			}
		}

		switch desc {
		case "true":
			selectStatementThreads += " ORDER BY created DESC"
		default:
			selectStatementThreads += " ORDER BY created ASC"
		}

		limit := c.Query("limit")
		if len(limit) > 0 {
			selectStatementThreads += " LIMIT " + limit
		}

		res := make([]ThreadStruct, 0)
		rows, err := db.Query(selectStatementThreads)
		switch err {
		case nil:
		case sql.ErrNoRows:
		default:
			panic(err)
		}
		for rows.Next() {
			var tts ThreadStruct
			err = rows.Scan(&tts.Id, &tts.Author, &tts.Title, &tts.Created, &tts.Message, &tts.Slug)
			tts.ForumSlug = forumSlug
			common.Check(err)
			res = append(res, tts)
		}
		content, _ := json.Marshal(res)
		return string(content), 200
	}
	var res common.ErrStruct
	res.Message = "Can't find forum with given slug!"
	content, _ := json.Marshal(res)
	return string(content), 404
}