package post

import (
	"github.com/go-ozzo/ozzo-routing"
	"encoding/json"
	"technoparkdb/common"
	"time"
	"technoparkdb/user"
	"technoparkdb/database"
	"strconv"
	"github.com/jackc/pgx"
	"strings"
)

type PostStruct struct {
	Id int `json:"id,omitempty"`
	ParentId int `json:"parent,omitempty"`
	AuthorId int `json:"author_id,omitempty"`
	AuthorName string `json:"author,omitempty"`
	Created time.Time `json:"created,omitempty"`
	ForumId int `json:"forum_id,omitempty"`
	ForumSlug string `json:"forum,omitempty"`
	Edited bool `json:"isEdited,omitempty"`
	Message string `json:"message,omitempty"`
	ThreadId int `json:"thread,omitempty"`
}

type ArrPostStruct []PostStruct
type CheckPost struct {
	AuthorId int
	AuthorName string
	PostParentId int
}

const insertStatement = "INSERT INTO posts (author_id, author_name, message, parent_id, thread_id, forum_id, forum_slug, created) VALUES ($1,$2,$3, $4, $5, $6, $7, $8) RETURNING created, id"
const selectStatement = "SELECT thread_id FROM posts WHERE id=$1"

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

func GetPostThread(id int) (int){
	db := database.DB
	var thread int
	row := db.QueryRow(selectStatement, id)
	err := row.Scan(&thread)
	switch err {
	case pgx.ErrNoRows:
		return -1
	case nil:
		return thread
	default:
		panic(err)
	}
}

func Create(c *routing.Context) (string, int) {
	db := database.DB
	POST := getArrayPost(c)
	defer c.Request.Body.Close()
	created := time.Now()
	createdTime := created.Format("2006-01-02T15:04:05.000000Z")
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
	if err == pgx.ErrNoRows {
		var res common.ErrStruct
		res.Message = "Can't found thread"
		content, _ := json.Marshal(res)
		return string(content), 404
	}

	var CheckPostArr []CheckPost
	for _, post := range POST {
		var ta CheckPost
		ta.AuthorId, ta.AuthorName = user.GetUserId(post.AuthorName)
		if ta.AuthorId < 0 {
			var res common.ErrStruct
			res.Message = "Can't found user who created post. Aborting"
			content, _ := json.Marshal(res)
			return string(content), 404
		}
		ta.PostParentId = post.ParentId
		parentThreadID := GetPostThread(post.ParentId)
		if ta.PostParentId > 0 && threadId != parentThreadID {
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
		row := db.QueryRow(insertStatement, authorId, author, message, parentId, threadId, forumId, forumSlug, createdTime)
		err := row.Scan(&tres.Created, &tres.Id)
		common.Check(err)

		tres.AuthorName = author
		tres.Message = message
		tres.ThreadId = threadId
		tres.ForumSlug = forumSlug
		tres.ParentId = parentId
		res = append(res, tres)
	}

	content, _ := json.Marshal(res)
	return string(content), 201
}

func Update(c *routing.Context) (string, int) {
	db := database.DB
	POST := getPost(c)
	defer c.Request.Body.Close()

	postID := c.Param("id")
	message := POST.Message
	var res PostStruct
	var statement string
	var err error
	statement = "SELECT message, author_name, created, forum_slug, id, thread_id FROM posts WHERE id=$1"
	row := db.QueryRow(statement, postID)
	err = row.Scan(&res.Message, &res.AuthorName, &res.Created, &res.ForumSlug, &res.Id, &res.ThreadId)
	switch err {
	case pgx.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Can't found post."
		content, _ := json.Marshal(res)
		return string(content), 404
	case nil:
		if len(message) > 0 && message != res.Message {
			statement = "UPDATE posts SET message=$1, is_edited=true WHERE id=$2"
			row := db.QueryRow(statement, message, postID)
			err = row.Scan()
			res.Edited = true
			res.Message = message
		}
		content, _ := json.Marshal(res)
		return string(content), 200
	default:
		panic(err)
	}
}

func Details(c *routing.Context) (string, int) {
	db := database.DB

	postID := c.Param("id")

	relatedGet := c.Query("related")
	relatedGetArr := strings.Split(relatedGet, ",")
	related := make(map[string]string)
	for _, rel := range relatedGetArr {
		related[rel] = rel
	}

	type ForumStruct struct {
		User string `json:"user,omitempty"`
		Title string `json:"title,omitempty"`
		Slug string `json:"slug,omitempty"`
		Posts int `json:"posts,omitempty"`
		Threads int `json:"threads,omitempty"`
	}
	type ThreadStruct struct {
		Id int `json:"id,omitempty"`
		Slug string `json:"slug,omitempty"`
		Author string `json:"author,omitempty"`
		Title string `json:"title,omitempty"`
		ForumSlug string `json:"forum,omitempty"`
		Message string `json:"message,omitempty"`
		Created time.Time `json:"created,omitempty"`
	}
	var result struct {
		Post PostStruct `json:"post"`
		Author *user.UserStruct `json:"author,omitempty"`
		Forum *ForumStruct `json:"forum,omitempty"`
		Thread *ThreadStruct `json:"thread,omitempty"`
	}

	selectStatement := `SELECT author_id, author_name, created, forum_slug, thread_id, is_edited, message, id FROM posts WHERE id=$1`
	row := db.QueryRow(selectStatement, postID)
	err := row.Scan(&result.Post.AuthorId, &result.Post.AuthorName, &result.Post.Created, &result.Post.ForumSlug, &result.Post.ThreadId, &result.Post.Edited, &result.Post.Message, &result.Post.Id)

	switch err {
	case pgx.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Can't found post."
		content, _ := json.Marshal(res)
		return string(content), 404
	case nil:
		if _, ok := related["user"]; ok {
			selectStatement = `SELECT about, email, fullname, nickname FROM users WHERE id=$1`
			row := db.QueryRow(selectStatement, result.Post.AuthorId)
			var tAuthor user.UserStruct
			err := row.Scan(&tAuthor.About, &tAuthor.Email, &tAuthor.Fullname, &tAuthor.Nickname)
			result.Author = &tAuthor
			switch err {
			case pgx.ErrNoRows:
				var res common.ErrStruct
				res.Message = "Can't found post author."
				content, _ := json.Marshal(res)
				return string(content), 404
			case nil:
			default:
				panic(err)
			}
		}
		if _, ok := related["forum"]; ok {
			selectStatement = `SELECT owner_nickname, title, slug, posts_count, threads_count FROM forums WHERE slug=$1`
			row := db.QueryRow(selectStatement, result.Post.ForumSlug)
			var tForum ForumStruct
			err := row.Scan(&tForum.User, &tForum.Title, &tForum.Slug, &tForum.Posts, &tForum.Threads)
			result.Forum = &tForum
			switch err {
			case pgx.ErrNoRows:
				var res common.ErrStruct
				res.Message = "Can't found post forum."
				content, _ := json.Marshal(res)
				return string(content), 404
			case nil:
			default:
				panic(err)
			}
		}
		if _, ok := related["thread"]; ok {
			selectStatement = `SELECT author_name, forum_slug, title, created, message, id, slug FROM threads WHERE id=$1`
			row := db.QueryRow(selectStatement, result.Post.ThreadId)
			var tThread ThreadStruct
			err := row.Scan(&tThread.Author, &tThread.ForumSlug, &tThread.Title, &tThread.Created, &tThread.Message, &tThread.Id, &tThread.Slug)
			result.Thread = &tThread
			switch err {
			case pgx.ErrNoRows:
				var res common.ErrStruct
				res.Message = "Can't found post thread."
				content, _ := json.Marshal(res)
				return string(content), 404
			case nil:
			default:
				panic(err)
			}
		}
		content, _ := json.Marshal(result)
		return string(content), 200
	default:
		panic(err)
	}
}