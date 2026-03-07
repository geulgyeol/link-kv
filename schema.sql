-- links table stores URLs and its last accessed timestamp (for re-crawling purposes)

CREATE TABLE IF NOT EXISTS links (
                       url text not null unique primary key,
                       last_enqueued_at timestamp not null default current_timestamp,
                       created_at timestamp not null default current_timestamp
);
CREATE INDEX IF NOT EXISTS idx_links_last_enqueued_at ON links (last_enqueued_at);
CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at);

-- blog_users table stores users of the blog platforms for re-crawling purposes

CREATE TABLE IF NOT EXISTS blog_users (
                            blog_platform text not null,
                            user_id text not null,
                            last_enqueued_at timestamp not null default current_timestamp,
                            created_at timestamp not null default current_timestamp,
                            primary key (blog_platform, user_id)
);

CREATE INDEX IF NOT EXISTS idx_blog_users_last_enqueued_at ON blog_users (last_enqueued_at);
CREATE INDEX IF NOT EXISTS idx_blog_users_created_at ON blog_users (created_at);

-- blog_posts table stores blog posts' path and metadata

CREATE TABLE IF NOT EXISTS blog_posts (
                            blog_platform text not null,
                            post_url text not null,
                            path text not null,
                            published_at timestamp,
                            created_at timestamp not null default current_timestamp,
                            updated_at timestamp not null default current_timestamp,
                            primary key (post_url)
);

CREATE INDEX IF NOT EXISTS idx_blog_posts_created_at ON blog_posts (created_at);
CREATE INDEX IF NOT EXISTS idx_blog_posts_updated_at ON blog_posts (updated_at);

-- task queues stores backlog of tasks to be processed

-- fetch content
CREATE TABLE IF NOT EXISTS content_queue (
                               id bigserial primary key,
                               payload text not null,
                               enqueued_at timestamp not null default current_timestamp,
                               locked_until timestamp,
                               attempts integer not null default 0,
                               status text not null default 'waiting' -- waiting, processing, done, failed
);


CREATE INDEX IF NOT EXISTS idx_content_queue_enqueued_at ON content_queue (enqueued_at);

-- fetch related profiles
CREATE TABLE IF NOT EXISTS profile_queue (
                               id bigserial primary key,
                               payload text not null,
                               enqueued_at timestamp not null default current_timestamp,
                               locked_until timestamp,
                               attempts integer not null default 0,
                               status text not null default 'waiting' -- waiting, processing, done, failed
);

CREATE INDEX IF NOT EXISTS idx_profile_queue_enqueued_at ON profile_queue (enqueued_at);

-- fetch user data
CREATE TABLE IF NOT EXISTS user_queue (
                            id bigserial primary key,
                            payload text not null,
                            enqueued_at timestamp not null default current_timestamp,
                            locked_until timestamp,
                            attempts integer not null default 0,
                            status text not null default 'waiting' -- waiting, processing, done, failed
);

CREATE INDEX IF NOT EXISTS idx_user_queue_enqueued_at ON user_queue (enqueued_at);
