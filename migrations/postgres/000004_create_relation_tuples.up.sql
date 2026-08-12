CREATE TABLE relation_tuples (
    id SERIAL PRIMARY KEY,
    object_id VARCHAR(255) NOT NULL,
    relation VARCHAR(100) NOT NULL,
    subject VARCHAR(255) NOT NULL,
    CONSTRAINT uq_relation_tuple UNIQUE (object_id, relation, subject)
);

CREATE INDEX idx_relation_lookup ON relation_tuples (object_id, relation);