CREATE TABLE subject_attributes (
    subject_id VARCHAR(255) NOT NULL,
    attr_key VARCHAR(100) NOT NULL,
    attr_value VARCHAR(500) NOT NULL,
    PRIMARY KEY (subject_id, attr_key)
);