package forum

import (
	"github.com/go-ozzo/ozzo-routing"
	"encoding/json"
	"technoparkdb/common"
	"technoparkdb/user"
	"technoparkdb/database"
	"github.com/jackc/pgx"
	"technoparkdb/thread"
	"strconv"
)

type ForumStruct struct {
	User string `json:"user,omitempty"`
	Title string `json:"title,omitempty"`
	Slug string `json:"slug,omitempty"`
	Posts int `json:"posts,omitempty"`
	Threads int `json:"threads,omitempty"`
}

const insertStatement = "INSERT INTO forums (owner_id, owner_nickname, title, slug) VALUES ($1,$2,$3,$4)"
const selectStatementSlug = "SELECT slug, title FROM forums WHERE slug=$1"
const selectStatementSlugAll = "SELECT slug, title, owner_nickname, posts_count, threads_count FROM forums WHERE slug=$1"

func getPost(c *routing.Context) ForumStruct {
	var POST ForumStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func Create(c *routing.Context) (string, int) {
	db := database.DB
	POST := getPost(c)
	defer c.Request.Body.Close()

	slug := POST.Slug
	title := POST.Title
	nick := POST.User
	userId := -1

	userId, nick = user.GetUserId(nick)
	if userId >= 0 {
		var res ForumStruct
		row := db.QueryRow(insertStatement, userId, nick, title, slug)
		err := row.Scan()
		if err != nil && err != pgx.ErrNoRows {
			row := db.QueryRow(selectStatementSlug, slug)
			err := row.Scan(&res.Slug, &res.Title)
			res.User = nick
			switch err {
			case nil:
				content, _ := json.Marshal(res)
				return string(content), 409
			default:
				panic(err)
			}
		}
		res.User = nick
		res.Slug = slug
		res.Title = title
		res.Threads = 0
		res.Posts = 0
		content, _ := json.Marshal(res)
		return string(content), 201
	}
	var res common.ErrStruct
	res.Message = "Can not create forum. User-creator not found"
	content, _ := json.Marshal(res)
	return string(content), 404
}

func Details(c *routing.Context) (string, int) {
	db := database.DB
	slug := c.Param("slug")
	var res ForumStruct

	row := db.QueryRow(selectStatementSlugAll, slug)
	err := row.Scan(&res.Slug, &res.Title, &res.User, &res.Posts, &res.Threads)
	switch err {
	case nil:
		content, _ := json.Marshal(res)
		return string(content), 200
	case pgx.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Forum not found!"
		content, _ := json.Marshal(res)
		return string(content), 404
	default:
		panic(err)
	}
}

func GetUsers(c *routing.Context) (string, int) {
	db := database.DB
	forumSlug := c.Param("slug")
	forumId, forumSlug := thread.GetForumSlugId(forumSlug)
	if forumId >= 0 {
		selectStatement := "SELECT about, email, fullname, nickname FROM users u JOIN forum_users fu ON (u.id = fu.user_id) WHERE fu.forum_id=" + strconv.Itoa(forumId)

		desc := c.Query("desc")
		since := c.Query("since")
		if len(since) > 0 {
			switch desc {
			case "true":
				selectStatement += " AND u.nickname < '" + since + "'"
			default:
				selectStatement += " AND u.nickname > '" + since + "'"
			}
		}

		switch desc {
		case "true":
			selectStatement += " ORDER BY u.nickname DESC"
		default:
			selectStatement += " ORDER BY u.nickname ASC"
		}

		limit := c.Query("limit")
		if len(limit) > 0 {
			selectStatement += " LIMIT " + limit
		}

		res := make([]user.UserStruct, 0)
		rows, err := db.Query(selectStatement)
		for rows.Next() {
			var tus user.UserStruct
			err = rows.Scan(&tus.About, &tus.Email, &tus.Fullname, &tus.Nickname)
			common.Check(err)
			res = append(res, tus)
		}
		content, _ := json.Marshal(res)
		return string(content), 200
	}
	var res common.ErrStruct
	res.Message = "Can't find forum with given slug!"
	content, _ := json.Marshal(res)
	return string(content), 404
}
