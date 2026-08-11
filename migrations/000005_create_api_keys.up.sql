CREATE TABLE api_keys (
    id INT IDENTITY PRIMARY KEY,
    key_hash NVARCHAR(64) NOT NULL UNIQUE,
    client_name NVARCHAR(255) NOT NULL,
    rate_limit_rpm INT NOT NULL DEFAULT 60,
    is_active BIT NOT NULL DEFAULT 1,
    created_at DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
);

CREATE INDEX idx_api_keys_hash ON api_keys (key_hash);