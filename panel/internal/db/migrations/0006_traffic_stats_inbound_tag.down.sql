DROP INDEX IF EXISTS idx_traffic_stats_node_inbound_recorded;
DROP INDEX IF EXISTS idx_traffic_stats_node_recorded;

-- SQLite cannot drop columns directly. Recreate the table without inbound_tag.
CREATE TABLE traffic_stats__old (
  id INTEGER PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  node_id INTEGER REFERENCES nodes(id),
  upload INTEGER DEFAULT 0,
  download INTEGER DEFAULT 0,
  recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO traffic_stats__old (id, user_id, node_id, upload, download, recorded_at)
  SELECT id, user_id, node_id, upload, download, recorded_at
  FROM traffic_stats;

DROP TABLE traffic_stats;

ALTER TABLE traffic_stats__old RENAME TO traffic_stats;
