package thread

import (
	"github.com/go-ozzo/ozzo-routing"
	"encoding/json"
	"technoparkdb/common"
	"technoparkdb/database"
	"github.com/jackc/pgx"
	"time"
	"strconv"
)

type ThreadStruct struct {
	Id int `json:"id,omitempty"`
	Slug string `json:"slug,omitempty"`
	Author string `json:"author,omitempty"`
	Title string `json:"title,omitempty"`
	ForumSlug string `json:"forum,omitempty"`
	Message string `json:"message,omitempty"`
	Created time.Time `json:"created,omitempty"`
	Votes int `json:"votes,omitempty"`
}

type VoteStruct struct {
	Vote int `json:"voice"`
	Nickname string `json:"nickname"`
}

const insertStatement = "INSERT INTO threads (author_id, author_name, forum_id, forum_slug, title, created, message, slug) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id"
const selectStatementSlug = "SELECT id, created, message, title, author_name, forum_slug, votes FROM threads WHERE slug=$1"
const selectStatementID = "SELECT id, created, message, title, author_name, forum_slug, votes FROM threads WHERE id=$1"
const selectStatementForumSlugId = "SELECT id, slug FROM forums WHERE slug=$1"

func getPost(c *routing.Context) ThreadStruct {
	var POST ThreadStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func getVotePost(c *routing.Context) VoteStruct {
	var POST VoteStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func GetForumSlugId(slug string) (int, string){
	db := database.DB
	row := db.QueryRow(selectStatementForumSlugId, slug)
	var id int
	err := row.Scan(&id, &slug)
	switch err {
	case pgx.ErrNoRows:
		return -1, ""
	case nil:
		return id, slug
	default:
		panic(err)
	}
}

func getForumAuthorsInfo(author, slug string) (int, string, int, string){
	db := database.DB
	selectStatement :="SELECT forum.*, author.* FROM (SELECT id, slug FROM forums WHERE slug=$1) as forum, (SELECT id, nickname FROM users WHERE nickname=$2) as author"
	row := db.QueryRow(selectStatement, slug, author)
	var forumID int
	var authorID int
	var forumSlug string
	var authorNickname string
	err := row.Scan(&forumID, &forumSlug, &authorID, &authorNickname)
	switch err {
	case pgx.ErrNoRows:
		return -1, "", -1, ""
	case nil:
		return forumID, forumSlug, authorID, authorNickname
	default:
		panic(err)
	}
}

func CreateThread(c *routing.Context) (string, int) {
	db := database.DB
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

	forumId, forumSlug, authorId, author = getForumAuthorsInfo(author, forumSlug)
	if authorId >= 0 && forumId >= 0 {
		var res ThreadStruct

		//костыль, чтобы вставлять в базу nil, а не пустую строку. мб позже попробовать с дефолтным параметром
		err := pgx.ErrNoRows
		if len(slug) > 0 {
			row := db.QueryRow(selectStatementSlug, slug)
			err = row.Scan(&res.Id, &res.Slug, &res.Created, &res.Message, &res.Title, &res.Author, &res.ForumSlug, &res.Votes)
		}

		switch err {
		case pgx.ErrNoRows:
			var row *pgx.Row
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
	var res common.ErrStruct
	res.Message = "Can not create thread. Thread author not found"
	content, _ := json.Marshal(res)
	return string(content), 404
}

func getThread(slug string) (error, ThreadStruct) {
	id, ierr := strconv.Atoi(slug)
	db := database.DB
	var res ThreadStruct
	var row *pgx.Row
	if ierr != nil {
		row = db.QueryRow(selectStatementSlug, slug)
	} else {
		row = db.QueryRow(selectStatementID, id)
	}

	err := row.Scan(&res.Id, &res.Created, &res.Message, &res.Title, &res.Author, &res.ForumSlug, &res.Votes)
	return err, res
}

func Details(c *routing.Context) (string, int) {
	slug := c.Param("slugid")
	err, res := getThread(slug)
	switch err {
	case pgx.ErrNoRows:
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

func Update(c *routing.Context) (string, int) {
	db := database.DB
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
	_, err := strconv.Atoi(slug)
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
		case pgx.ErrNoRows:
			var resErr common.ErrStruct
			resErr.Message = "Thread not found!"
			content, _ := json.Marshal(resErr)
			return string(content), 404
		case nil:
			content, _ := json.Marshal(resOk)
			return string(content), 200
		}
	}
	err, res := getThread(slug)
	switch err {
	case pgx.ErrNoRows:
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

func GetThreads(c *routing.Context) (string, int) {
	db := database.DB
	forumSlug := c.Param("slug")
	forumId, forumSlug := GetForumSlugId(forumSlug)
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

func Vote(c *routing.Context) (string, int) {
	db := database.DB
	POST := getVotePost(c)
	defer c.Request.Body.Close()
	nickname := POST.Nickname
	voice := POST.Vote
	slug := c.Param("slugid")
	slugId, ierr := strconv.Atoi(slug)
	transaction, _ := db.Begin()

	if ierr != nil {
		selectStatement := "SELECT id FROM threads WHERE slug=$1"
		err := transaction.QueryRow(selectStatement, slug).Scan(&slugId)
		if err == pgx.ErrNoRows {
			var resErr common.ErrStruct
			resErr.Message = "Thread or user not found!"
			content, _ := json.Marshal(resErr)
			transaction.Rollback()
			return string(content), 404
		}
	}

	id := -1
	var insertStatement string
	insertStatement = "INSERT INTO thread_votes (user_nickname, thread_id, vote) VALUES ($1,$2,$3) ON CONFLICT (user_nickname, thread_id) DO UPDATE SET vote=$3 RETURNING id"
	transaction.QueryRow(insertStatement, nickname, slugId, voice).Scan(&id)

	if id < 0 {
		transaction.Rollback()
		var resErr common.ErrStruct
		resErr.Message = "Thread or user not found!"
		content, _ := json.Marshal(resErr)
		return string(content), 404
	} else {
		ids, ierr := strconv.Atoi(slug)
		var res ThreadStruct
		var row *pgx.Row
		if ierr != nil {
			row = transaction.QueryRow(selectStatementSlug, slug)
		} else {
			row = transaction.QueryRow(selectStatementID, ids)
		}

		row.Scan(&res.Id, &res.Created, &res.Message, &res.Title, &res.Author, &res.ForumSlug, &res.Votes)

		content, _ := json.Marshal(res)
		transaction.Commit()
		return string(content), 200
	}
}
