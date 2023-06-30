CREATE IF NOT EXISTS TABLE {{.EntryType}}(
    List LowCardinality(String),
    Index String,
    ExtraData Map(String, String) DEFAULT {}
) ENGINE = ReplacingMergeTree()
ORDER BY (
    List, Index
) PRIMARY KEY (
    List, Index
)