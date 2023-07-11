CREATE TABLE IF NOT EXISTS {{.EntryType}}(
    List LowCardinality(String),
    Idx String,
    ExtraData Map(String, String)
) ENGINE = ReplacingMergeTree()
ORDER BY (
    List, Idx
) PRIMARY KEY (
    List, Idx
)