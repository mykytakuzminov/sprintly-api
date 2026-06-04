CREATE TABLE IF NOT EXISTS boards (
  id          UUID        PRIMARY KEY,
  owner_id    UUID        NOT NULL REFERENCES users(id),
  name        TEXT        NOT NULL,
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS columns (
  id       UUID PRIMARY KEY,
  board_id UUID NOT NULL REFERENCES boards(id),
  name     TEXT NOT NULL,
  position INT  NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
  id          UUID        PRIMARY KEY,
  owner_id    UUID        NOT NULL REFERENCES users(id),
  column_id   UUID        NOT NULL REFERENCES columns(id),
  assignee_id UUID        REFERENCES users(id),
  name        TEXT        NOT NULL,
  description TEXT,
  due_date    TIMESTAMPTZ,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
