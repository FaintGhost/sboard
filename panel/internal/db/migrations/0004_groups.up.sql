CREATE TABLE groups (
  id INTEGER PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  description TEXT,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_groups (
  user_id INTEGER REFERENCES users(id),
  group_id INTEGER REFERENCES groups(id),
  PRIMARY KEY (user_id, group_id)
);

ALTER TABLE nodes ADD COLUMN group_id INTEGER REFERENCES groups(id);
