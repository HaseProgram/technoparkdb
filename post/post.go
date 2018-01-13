package post

import (
	"github.com/go-ozzo/ozzo-routing"
	"encoding/json"
	"github.com/HaseProgram/technoparkdb/common"
	"time"
	"github.com/HaseProgram/technoparkdb/user"
	"github.com/HaseProgram/technoparkdb/database"
	"strconv"
	"github.com/jackc/pgx"
	"strings"
	"fmt"
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

	selectThreadStatement := "SELECT id, forum_id, forum_slug, slug from threads WHERE"
	if err == nil {
		selectThreadStatement += " id=" + threadSlugId
	} else {
		selectThreadStatement += " slug='" + threadSlugId + "'"
	}

	var threadId int
	var threadSlug string
	var forumId int
	var forumSlug string

	transaction, _ := db.Begin()

	row := transaction.QueryRow(selectThreadStatement)
	err = row.Scan(&threadId, &forumId, &forumSlug, &threadSlug)
	if err == pgx.ErrNoRows {
		transaction.Rollback()
		var res common.ErrStruct
		res.Message = "Can't found thread"
		content, _ := json.Marshal(res)
		return string(content), 404
	}

	var CheckPostArr []CheckPost
	transaction.Prepare("select_user", "SELECT id, nickname FROM users WHERE nickname=$1")
	transaction.Prepare("get_parent","SELECT thread_id FROM posts WHERE id=$1")

	//batch := db.BeginBatch()
	for _, post := range POST {
		var ta CheckPost

		ta.AuthorId = -1

		row := transaction.QueryRow("select_user", post.AuthorName)
		err := row.Scan(&ta.AuthorId, &ta.AuthorName)

		if ta.AuthorId < 0 {
			transaction.Rollback()
			var res common.ErrStruct
			res.Message = "Can't found user who created post. Aborting"
			content, _ := json.Marshal(res)
			return string(content), 404
		}
		ta.PostParentId = post.ParentId
		var parentThreadID int
		row = transaction.QueryRow("get_parent", post.ParentId)
		err = row.Scan(&parentThreadID)
		switch err {
		case pgx.ErrNoRows:
			parentThreadID = -1
		case nil:
		default:
			panic(err)
		}
		if ta.PostParentId > 0 && threadId != parentThreadID {
			transaction.Rollback()
			var res common.ErrStruct
			res.Message = "Can't found parent post. Aborting"
			content, _ := json.Marshal(res)
			return string(content), 409
		}

		CheckPostArr = append(CheckPostArr, ta)
	}

	res := make([]PostStruct, 0)

	transaction.Prepare("insert_post", insertStatement)

	for index, post := range POST {
		var tres PostStruct
		message := post.Message
		author := CheckPostArr[index].AuthorName
		authorId := CheckPostArr[index].AuthorId
		parentId := CheckPostArr[index].PostParentId
		row := transaction.QueryRow("insert_post", authorId, author, message, parentId, threadId, forumId, forumSlug, createdTime)

		err := row.Scan(&tres.Created, &tres.Id)
		common.Check(err)

		tres.AuthorName = author
		tres.Message = message
		tres.ThreadId = threadId
		tres.ForumSlug = forumSlug
		tres.ParentId = parentId
		res = append(res, tres)
	}

	transaction.Commit()
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
	//t0 := time.Now()
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
		Votes int `json:"votes,omitempty"`
	}
	var result struct {
		Post PostStruct `json:"post"`
		Author *user.UserStruct `json:"author,omitempty"`
		Forum *ForumStruct `json:"forum,omitempty"`
		Thread *ThreadStruct `json:"thread,omitempty"`
	}

	selectStatement := `SELECT author_id, author_name, created, forum_slug, thread_id, is_edited, message, id, parent_id FROM posts WHERE id=$1`
	row := db.QueryRow(selectStatement, postID)
	err := row.Scan(&result.Post.AuthorId, &result.Post.AuthorName, &result.Post.Created, &result.Post.ForumSlug, &result.Post.ThreadId, &result.Post.Edited, &result.Post.Message, &result.Post.Id, &result.Post.ParentId)

	switch err {
	case pgx.ErrNoRows:
		var res common.ErrStruct
		res.Message = "Can't found post."
		content, _ := json.Marshal(res)
		//t1 := time.Now();
		//fmt.Println("Get post details (no post): ", t1.Sub(t0), "404");
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
				//t1 := time.Now();
				//fmt.Println("Get post details (no author): ", t1.Sub(t0), "404", relatedGet);
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
				//t1 := time.Now();
				//fmt.Println("Get post details (no forum): ", t1.Sub(t0), "404", relatedGet);
				return string(content), 404
			case nil:
			default:
				panic(err)
			}
		}
		if _, ok := related["thread"]; ok {
			selectStatement = `SELECT author_name, forum_slug, title, created, message, id, slug, votes FROM threads WHERE id=$1`
			row := db.QueryRow(selectStatement, result.Post.ThreadId)
			var tThread ThreadStruct
			err := row.Scan(&tThread.Author, &tThread.ForumSlug, &tThread.Title, &tThread.Created, &tThread.Message, &tThread.Id, &tThread.Slug, &tThread.Votes)
			result.Thread = &tThread
			if err == pgx.ErrNoRows {
				var res common.ErrStruct
				res.Message = "Can't found post thread."
				content, _ := json.Marshal(res)
				//t1 := time.Now();
				//fmt.Println("Get post details (no thread): ", t1.Sub(t0), "404", relatedGet);
				return string(content), 404
			}
		}
		content, _ := json.Marshal(result)
		//t1 := time.Now();
		//fmt.Println("Get post details: ", t1.Sub(t0), "200", relatedGet);
		return string(content), 200
	default:
		panic(err)
	}
}

func GetPosts(c *routing.Context) (string, int) {
	t0 := time.Now()
	db := database.DB

	threadSlugId := c.Param("slugid")
	_, err := strconv.Atoi(threadSlugId)

	selectThreadStatement := "SELECT id, slug FROM threads WHERE"
	if err == nil {
		selectThreadStatement += " id=" + threadSlugId
	} else {
		selectThreadStatement += " slug='" + threadSlugId + "'"
	}

	var threadId int
	var threadSlug string

	row := db.QueryRow(selectThreadStatement)
	err = row.Scan(&threadId, &threadSlug)
	if err == pgx.ErrNoRows {
		var res common.ErrStruct
		res.Message = "Can't found thread"
		content, _ := json.Marshal(res)
		return string(content), 404
	}

	limit := c.Query("limit")
	since := c.Query("since")
	sort := c.Query("sort")
	desc := c.Query("desc")

	selectStatement := "SELECT created, id, is_edited, message, parent_id, author_id, thread_id, forum_slug, forum_id, author_name FROM posts WHERE"

	switch sort {
	case "tree":
		selectStatement += " thread_id=" + strconv.Itoa(threadId)
		if len(since) > 0 {
			switch desc {
			case "true":
				selectStatement += " AND path_to_post < (SELECT path_to_post FROM posts WHERE id=" + since + ")"
			default:
				selectStatement += " AND path_to_post > (SELECT path_to_post FROM posts WHERE id=" + since + ")"
			}
		}
		switch desc {
		case "true":
			selectStatement += " ORDER BY path_to_post DESC"
		default:
			selectStatement += " ORDER BY path_to_post ASC"
		}
		if len(limit) > 0 {
			selectStatement += " LIMIT " + limit
		}
	case "parent_tree":
		selectStatement += " rootidx IN (SELECT id FROM posts WHERE thread_id=" + strconv.Itoa(threadId) + " AND parent_id=0"
		if len(since) > 0 {
			switch desc {
			case "true":
				selectStatement += " AND path_to_post < (SELECT path_to_post FROM posts WHERE id=" + since + ")"
			default:
				selectStatement += " AND path_to_post > (SELECT path_to_post FROM posts WHERE id=" + since + ")"
			}
		}
		if len(limit) > 0 {
			limit = " LIMIT " + limit
		}
		switch desc {
		case "true":
			selectStatement += " ORDER BY id DESC" + limit + ") ORDER BY path_to_post DESC"
		default:
			selectStatement += " ORDER BY id ASC" + limit + ") ORDER BY path_to_post ASC"
		}
	default: // flat
		selectStatement += " thread_id=" + strconv.Itoa(threadId)
		if len(since) > 0 {
			switch desc {
			case "true":
				selectStatement += " AND id < '" + since + "'"
			default:
				selectStatement += " AND id > '" + since + "'"
			}
		}
		switch desc {
		case "true":
			selectStatement += " ORDER BY id DESC"
		default:
			selectStatement += " ORDER BY id ASC"
		}
		if len(limit) > 0 {
			selectStatement += " LIMIT " + limit
		}
	}

	res := make([]PostStruct, 0)
	rows, err := db.Query(selectStatement)
	for rows.Next() {
		var tps PostStruct
		err = rows.Scan(&tps.Created, &tps.Id, &tps.Edited, &tps.Message, &tps.ParentId, &tps.AuthorId, &tps.ThreadId, &tps.ForumSlug, &tps.ForumId, &tps.AuthorName)
		common.Check(err)
		res = append(res, tps)
	}
	content, _ := json.Marshal(res)
	t1 := time.Now();
	fmt.Println("Get posts: ", t1.Sub(t0), selectStatement);
	return string(content), 200
}