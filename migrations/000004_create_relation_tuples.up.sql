CREATE TABLE relation_tuples (
    id INT IDENTITY PRIMARY KEY,
    object_id NVARCHAR(255) NOT NULL,
    relation NVARCHAR(100) NOT NULL,
    subject NVARCHAR(255) NOT NULL,
    CONSTRAINT UQ_relation_tuple UNIQUE (object_id, relation, subject)
);

CREATE INDEX idx_relation_lookup ON relation_tuples (object_id, relation);