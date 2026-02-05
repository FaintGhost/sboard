ALTER TABLE nodes ADD COLUMN api_address TEXT;
ALTER TABLE nodes ADD COLUMN api_port INTEGER;
ALTER TABLE nodes ADD COLUMN public_address TEXT;

ALTER TABLE inbounds ADD COLUMN public_port INTEGER;

CREATE TABLE user_nodes (
  user_id INTEGER REFERENCES users(id),
  node_id INTEGER REFERENCES nodes(id),
  PRIMARY KEY (user_id, node_id)
);
