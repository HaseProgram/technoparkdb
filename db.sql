SET SYNCHRONOUS_COMMIT = 'off';

DROP TABLE forums;
DROP TABLE threads;
DROP TABLE posts;
DROP TABLE forum_users;
DROP TABLE thread_votes;
DROP TABLE users;

CREATE TABLE IF NOT EXISTS users (
	id SERIAL PRIMARY KEY,
	about TEXT DEFAULT NULL,
	email CITEXT UNIQUE,
	fullname TEXT DEFAULT NULL,
	nickname CITEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS users_nickname_uindex ON users (nickname);
CREATE INDEX IF NOT EXISTS users_id_nickname_index ON users (id, nickname);

CREATE TABLE IF NOT EXISTS forums (
	id SERIAL PRIMARY KEY,
	owner_id INTEGER NOT NULL,
	owner_nickname CITEXT,
	title TEXT NOT NULL,
	slug CITEXT NOT NULL,
	posts_count INTEGER DEFAULT 0,
	threads_count INTEGER DEFAULT 0
);

CREATE UNIQUE INDEX IF NOT EXISTS forums_slug_uindex ON forums (slug);

CREATE TABLE IF NOT EXISTS threads (
	id SERIAL PRIMARY KEY,
	author_id INTEGER NOT NULL,
	author_name CITEXT,
	forum_id INTEGER NOT NULL,
	forum_slug CITEXT,
	title TEXT  NOT NULL,
	created TIMESTAMPTZ DEFAULT NOW(),
	message TEXT DEFAULT NULL,
	votes INTEGER DEFAULT 0,
	slug CITEXT DEFAULT NULL :: CITEXT
);

CREATE INDEX IF NOT EXISTS threads_slug_index ON threads (slug);
CREATE INDEX IF NOT EXISTS threads_forum_slug_index ON threads (forum_slug);
CREATE INDEX IF NOT EXISTS threads_forum_id_index ON threads (forum_id);
CREATE INDEX IF NOT EXISTS threads_created_index ON threads (created);	
CREATE INDEX IF NOT EXISTS threads_forum_created_index ON threads (forum_slug, created);		


CREATE TABLE IF NOT EXISTS posts (
	id SERIAL PRIMARY KEY,
	parent_id INTEGER DEFAULT 0,
	author_id INTEGER NOT NULL,
	author_name CITEXT,
	created TIMESTAMPTZ DEFAULT NOW(),
	forum_id INTEGER NOT NULL,
	forum_slug CITEXT,
	is_edited BOOLEAN DEFAULT FALSE,
	message TEXT DEFAULT NULL,
	thread_id INTEGER NOT NULL,
	path_to_post INTEGER [],
	rootidx INTEGER
);

CREATE INDEX IF NOT EXISTS posts_forum_slug_index ON posts (id, forum_slug);
CREATE INDEX IF NOT EXISTS posts_forum_id_index ON posts (id, forum_id);
CREATE INDEX IF NOT EXISTS posts_parent_id_index ON posts (parent_id);
CREATE INDEX IF NOT EXISTS posts_thread_id_index ON posts (thread_id);
CREATE INDEX IF NOT EXISTS posts_thread_id_id_index ON posts (thread_id, id);
CREATE INDEX IF NOT EXISTS posts_root_index ON posts (rootidx);
--CREATE INDEX IF NOT EXISTS posts_thread_id_path_index ON posts (thread_id, path_to_post);
--CREATE INDEX IF NOT EXISTS posts_thread_id_parent_id_index ON posts (thread_id, parent_id);
--CREATE INDEX IF NOT EXISTS posts_thread_id_parent_id_path_index ON posts (thread_id, parent_id, path_to_post);

CREATE TABLE IF NOT EXISTS forum_users (
	user_id INTEGER NOT NULL,
	forum_id INTEGER NOT NULL
	--CONSTRAINT user_forum UNIQUE (user_id, forum_id)
);

CREATE INDEX IF NOT EXISTS forum_users_forum ON forum_users (forum_id);
CREATE UNIQUE INDEX IF NOT EXISTS forum_users_id_id ON forum_users (user_id, forum_id);

CREATE TABLE IF NOT EXISTS thread_votes (
	id SERIAL PRIMARY KEY,
	user_nickname CITEXT REFERENCES users (nickname) NOT NULL,
	thread_id INTEGER NOT NULL,
	--CONSTRAINT user_thread UNIQUE (user_nickname, thread_id),
	vote INTEGER
);

CREATE UNIQUE INDEX IF NOT EXISTS voyes_thread_user_uindex ON thread_votes (user_nickname, thread_id);

CREATE OR REPLACE FUNCTION update_thread_func() RETURNS TRIGGER AS
$update_thread_trig$
	BEGIN
		UPDATE forums SET threads_count = threads_count + 1 WHERE id = NEW.forum_id;
		INSERT INTO forum_users (user_id, forum_id) (SELECT NEW.author_id, NEW.forum_id) ON CONFLICT (user_id, forum_id) DO NOTHING;
		RETURN NEW;
	END;
$update_thread_trig$
LANGUAGE plpgsql;


CREATE OR REPLACE FUNCTION update_posts_func() RETURNS TRIGGER AS
$update_posts_trig$
	BEGIN
		UPDATE forums SET posts_count = posts_count + 1 WHERE id = NEW.forum_id;
		INSERT INTO forum_users (user_id, forum_id) (SELECT NEW.author_id, NEW.forum_id) ON CONFLICT (user_id, forum_id) DO NOTHING;
		RETURN NEW;
	END;
$update_posts_trig$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION path_posts_func() RETURNS TRIGGER AS
$path_posts_trig$
	BEGIN
		IF (NEW.parent_id = 0)
			THEN 
				NEW.path_to_post = ARRAY[NEW.id];
				NEW.rootidx = NEW.id;
			ELSE 
				NEW.path_to_post = (SELECT p.path_to_post || NEW.id FROM posts p WHERE id = NEW.parent_id);
				NEW.rootidx = NEW.path_to_post[1];
		END IF;
		UPDATE forums SET posts_count = posts_count + 1 WHERE id = NEW.forum_id;
		INSERT INTO forum_users (user_id, forum_id) (SELECT NEW.author_id, NEW.forum_id) ON CONFLICT (user_id, forum_id) DO NOTHING;
		RETURN NEW;
	END;
$path_posts_trig$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION vote_insert_func() RETURNS TRIGGER AS
$vote_insert_trig$
	BEGIN
		UPDATE threads SET votes=votes+NEW.vote WHERE id=NEW.thread_id;
		RETURN NEW;
	END;
$vote_insert_trig$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION vote_update_func() RETURNS TRIGGER AS
$vote_update_trig$
	BEGIN
		IF (NEW.vote!=OLD.vote) THEN
			IF (NEW.vote>0)
				THEN UPDATE threads SET votes=votes+2 WHERE id=NEW.thread_id;
				ELSE UPDATE threads SET votes=votes-2 WHERE id=NEW.thread_id;
			END IF;
		END IF;
		RETURN NEW;
	END;
$vote_update_trig$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_thread_trig ON threads;
DROP TRIGGER IF EXISTS update_posts_trig ON posts;
DROP TRIGGER IF EXISTS path_posts_trig ON posts;
DROP TRIGGER IF EXISTS vote_insert_trig ON thread_votes;
DROP TRIGGER IF EXISTS vote_update_trig ON thread_votes;
CREATE TRIGGER update_thread_trig AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE update_thread_func();
--CREATE TRIGGER update_posts_trig AFTER INSERT ON posts FOR EACH ROW EXECUTE PROCEDURE update_posts_func();
CREATE TRIGGER path_posts_trig BEFORE INSERT ON posts FOR EACH ROW EXECUTE PROCEDURE path_posts_func();
CREATE TRIGGER vote_insert_trig AFTER INSERT ON thread_votes FOR EACH ROW EXECUTE PROCEDURE vote_insert_func();
CREATE TRIGGER vote_update_trig AFTER UPDATE ON thread_votes FOR EACH ROW EXECUTE PROCEDURE vote_update_func();