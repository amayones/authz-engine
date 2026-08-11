CREATE TABLE subject_attributes (
    subject_id NVARCHAR(255) NOT NULL,
    attr_key NVARCHAR(100) NOT NULL,
    attr_value NVARCHAR(500) NOT NULL,
    PRIMARY KEY (subject_id, attr_key)
);