CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  about TEXT DEFAULT NULL,
  email CITEXT UNIQUE,
  fullname TEXT DEFAULT NULL,
  nickname CITEXT COLLATE ucs_basic UNIQUE
);

CREATE TABLE IF NOT EXISTS forums (
  id SERIAL PRIMARY KEY,
  owner_id INTEGER REFERENCES users (id) ON DELETE CASCADE NOT NULL,
  owner_nickname CITEXT,
  title TEXT NOT NULL,
  slug CITEXT UNIQUE NOT NULL,
  posts_count INTEGER DEFAULT 0,
  threads_count INTEGER DEFAULT 0
);

CREATE TABLE IF NOT EXISTS threads (
  id SERIAL PRIMARY KEY,
  author_id INTEGER REFERENCES users (id) ON DELETE CASCADE  NOT NULL,
  author_name CITEXT,
  forum_id INTEGER REFERENCES forums (id) ON DELETE CASCADE NOT NULL,
  forum_slug CITEXT,
  title TEXT  NOT NULL,
  created TIMESTAMPTZ DEFAULT NOW(),
  message TEXT DEFAULT NULL,
  votes INTEGER DEFAULT 0,
  slug CITEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  parent_id INTEGER DEFAULT 0,
  author_id INTEGER REFERENCES users (id) ON DELETE CASCADE   NOT NULL,
  author_name CITEXT,
  created TIMESTAMPTZ DEFAULT NOW(),
  forum_id INTEGER REFERENCES forums (id) ON DELETE CASCADE  NOT NULL,
  forum_slug CITEXT,
  is_edited BOOLEAN DEFAULT FALSE,
  message TEXT DEFAULT NULL,
  thread_id INTEGER REFERENCES threads (id) ON DELETE CASCADE NOT NULL,
  path_to_post INTEGER []
);

CREATE OR REPLACE FUNCTION update_thread_count_func() RETURNS TRIGGER AS
$update_thread_count_trig$
	BEGIN
		UPDATE forums SET threads_count = threads_count + 1 WHERE id = NEW.forum_id;
		RETURN NEW;
	END;
$update_thread_count_trig$
LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_posts_count_func() RETURNS TRIGGER AS
$update_posts_count_trig$
	BEGIN
		UPDATE forums SET posts_count = posts_count + 1 WHERE id = NEW.forum_id;
		RETURN NEW;
	END;
$update_posts_count_trig$
LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_thread_count_trig ON threads;
DROP TRIGGER IF EXISTS update_posts_count_trig ON posts;
CREATE TRIGGER update_thread_count_trig AFTER INSERT ON threads FOR EACH ROW EXECUTE PROCEDURE update_thread_count_func();
CREATE TRIGGER update_posts_count_trig AFTER INSERT ON posts FOR EACH ROW EXECUTE PROCEDURE update_posts_count_func();