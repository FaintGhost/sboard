CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  username TEXT UNIQUE NOT NULL,
  password_hash TEXT,
  traffic_limit INTEGER DEFAULT 0,
  traffic_used INTEGER DEFAULT 0,
  traffic_reset_day INTEGER DEFAULT 0,
  expire_at DATETIME,
  status TEXT DEFAULT 'active',
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE nodes (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  address TEXT NOT NULL,
  port INTEGER NOT NULL,
  secret_key TEXT NOT NULL,
  status TEXT DEFAULT 'offline',
  last_seen_at DATETIME,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE inbounds (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  tag TEXT NOT NULL,
  node_id INTEGER REFERENCES nodes(id),
  protocol TEXT NOT NULL,
  listen_port INTEGER NOT NULL,
  settings JSON NOT NULL,
  tls_settings JSON,
  transport_settings JSON,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(node_id, tag)
);

CREATE TABLE user_inbounds (
  user_id INTEGER REFERENCES users(id),
  inbound_id INTEGER REFERENCES inbounds(id),
  PRIMARY KEY (user_id, inbound_id)
);

CREATE TABLE traffic_stats (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  node_id INTEGER REFERENCES nodes(id),
  upload INTEGER DEFAULT 0,
  download INTEGER DEFAULT 0,
  recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
