-- SQLite doesn't support dropping columns easily.
-- We keep nodes.group_id as-is on down migration to avoid data loss / complex table rebuild.

DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS groups;

