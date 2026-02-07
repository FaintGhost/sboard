ALTER TABLE traffic_stats ADD COLUMN inbound_tag TEXT;

CREATE INDEX IF NOT EXISTS idx_traffic_stats_node_inbound_recorded
  ON traffic_stats(node_id, inbound_tag, recorded_at);

CREATE INDEX IF NOT EXISTS idx_traffic_stats_node_recorded
  ON traffic_stats(node_id, recorded_at);
