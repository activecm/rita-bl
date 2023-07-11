SELECT 
    List,
    Idx,
    ExtraData
FROM 
    {{.EntryType}}
WHERE
    Idx=@index