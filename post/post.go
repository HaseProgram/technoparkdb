package post

import (
	"github.com/go-ozzo/ozzo-routing"
	"database/sql"
	"encoding/json"
	"technoparkdb/common"
	"time"
	"technoparkdb/user"
	"strconv"
)

type PostStruct struct {
	Id int `json:"id"`
	ParentId int `json:"parent_id"`
	AuthorId int `json:"author_id"`
	AuthorName string `json:"author"`
	Created time.Time `json:"created"`
	ForumId int `json:"forum_id"`
	ForumSlug string `json:"forum"`
	Edited bool `json:"edited"`
	Message string `json:"message"`
	ThreadId int `json:"thread"`
}

type ArrPostStruct []PostStruct
type CheckPost struct {
	AuthorId int
	AuthorName string
	PostParentId int
}

const insertStatement = "INSERT INTO posts (author_id, author_name, message, parent_id, thread_id, forum_id, forum_slug) VALUES ($1,$2,$3, $4, $5, $6, $7) RETURNING created, id"
const selectStatement = "SELECT parent_id FROM posts WHERE id=$1"

func getPost(c *routing.Context) PostStruct {
	var POST PostStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func getArrayPost(c *routing.Context) ArrPostStruct {
	var POST ArrPostStruct
	c.Request.ParseForm()
	decoder := json.NewDecoder(c.Request.Body)
	err := decoder.Decode(&POST)
	common.Check(err)
	return POST
}

func GetPostId(id int, db *sql.DB) (int){
	row := db.QueryRow(selectStatement, id)
	err := row.Scan(&id)
	switch err {
	case sql.ErrNoRows:
		return -1
	case nil:
		return id
	default:
		panic(err)
	}
}

func Create(c *routing.Context, db *sql.DB) (string, int) {
	POST := getArrayPost(c)
	defer c.Request.Body.Close()

	threadSlugId := c.Param("slugid")
	_, err := strconv.Atoi(threadSlugId)

	selectThreadStatement := "SELECT id, slug, forum_id, forum_slug from threads WHERE"
	if err == nil {
		selectThreadStatement += " id=" + threadSlugId
	} else {
		selectThreadStatement += " slug='" + threadSlugId + "'"
	}

	var threadId int
	var threadSlug string
	var forumId int
	var forumSlug string

	row := db.QueryRow(selectThreadStatement)
	err = row.Scan(&threadId, &threadSlug, &forumId, &forumSlug)
	if err == sql.ErrNoRows {
		var res common.ErrStruct
		res.Message = "Can't found thread"
		content, _ := json.Marshal(res)
		return string(content), 404
	}

	var CheckPostArr []CheckPost
	for _, post := range POST {
		var ta CheckPost
		ta.AuthorId, ta.AuthorName = user.GetUserId(post.AuthorName, db)
		if ta.AuthorId < 0 {
			var res common.ErrStruct
			res.Message = "Can't found user who created post. Aborting"
			content, _ := json.Marshal(res)
			return string(content), 409
		}
		ta.PostParentId = post.ParentId
		if ta.PostParentId > 0 && GetPostId(post.ParentId, db) < 0 {
			var res common.ErrStruct
			res.Message = "Can't found parent post. Aborting"
			content, _ := json.Marshal(res)
			return string(content), 409
		}

		CheckPostArr = append(CheckPostArr, ta)
	}

	res := make([]PostStruct, 0)

	for index, post := range POST {
		var tres PostStruct
		message := post.Message
		author := CheckPostArr[index].AuthorName
		authorId := CheckPostArr[index].AuthorId
		parentId := CheckPostArr[index].PostParentId
		row := db.QueryRow(insertStatement, authorId, author, message, parentId, threadId, forumId, forumSlug)
		err := row.Scan(&tres.Created, &tres.Id)
		common.Check(err)

		tres.AuthorName = author
		tres.Message = message
		tres.ThreadId = threadId
		tres.ForumSlug = forumSlug
		res = append(res, tres)
	}

	content, _ := json.Marshal(res)
	return string(content), 201
}