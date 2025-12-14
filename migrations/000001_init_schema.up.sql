CREATE TABLE IF NOT EXISTS blacklist (
	ip_range cidr NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_ip_range_black_list ON blacklist(ip_range);

CREATE TABLE IF NOT EXISTS whitelist (
	ip_range cidr NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_ip_range_white_list ON whitelist(ip_range);
