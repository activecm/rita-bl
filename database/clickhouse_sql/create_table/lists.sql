CREATE IF NOT EXISTS TABLE Lists(
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