-- +goose Up
CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       name TEXT NOT NULL,
                       email TEXT NOT NULL UNIQUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE projects (
                          id BIGSERIAL PRIMARY KEY,
                          name TEXT NOT NULL,
                          description TEXT,
                          status TEXT NOT NULL DEFAULT 'active',
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE project_members (
                                 project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                                 user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                 role TEXT NOT NULL DEFAULT 'member',
                                 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                 PRIMARY KEY (project_id, user_id)
);

CREATE TABLE tasks (
                       id BIGSERIAL PRIMARY KEY,
                       project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
                       assignee_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
                       title TEXT NOT NULL,
                       description TEXT,
                       status TEXT NOT NULL DEFAULT 'new',
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE comments (
                          id BIGSERIAL PRIMARY KEY,
                          task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
                          author_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          text TEXT NOT NULL,
                          created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE task_history (
                              id BIGSERIAL PRIMARY KEY,
                              task_id BIGINT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
                              old_status TEXT NOT NULL,
                              new_status TEXT NOT NULL,
                              changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX idx_comments_task_id ON comments(task_id);
CREATE INDEX idx_task_history_task_id ON task_history(task_id);

-- +goose Down
DROP TABLE IF EXISTS task_history;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;