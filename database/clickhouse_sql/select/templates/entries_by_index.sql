SELECT 
    List,
    Index,
    ExtraData
FROM 
    {{.EntryType}}
WHERE
    Index=@index