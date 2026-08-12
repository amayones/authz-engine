CREATE TABLE subject_roles (
    subject_id VARCHAR(255) NOT NULL,
    role_name VARCHAR(100) NOT NULL,
    PRIMARY KEY (subject_id, role_name),
    FOREIGN KEY (role_name) REFERENCES roles(name) ON DELETE CASCADE
);