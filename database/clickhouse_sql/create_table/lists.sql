CREATE TABLE IF NOT EXISTS Lists(
    Name LowCardinality(String),
    Types Array(LowCardinality(String)),
    LastCachedAt DateTime,
    CacheDuration Int64
) ENGINE = ReplacingMergeTree()
ORDER BY (
    Name
) PRIMARY KEY (
    Name
)