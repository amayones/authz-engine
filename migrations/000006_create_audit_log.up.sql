CREATE TABLE audit_log (
    id           INT IDENTITY PRIMARY KEY,
    actor        NVARCHAR(255) NOT NULL,      -- siapa yang melakukan aksi (client_name dari API key)
    action       NVARCHAR(100) NOT NULL,      -- misal "assign_role", "write_relation"
    target       NVARCHAR(500) NOT NULL,      -- objek yang terpengaruh, misal "user:alice -> editor"
    detail       NVARCHAR(MAX) NULL,          -- JSON detail tambahan
    occurred_at  DATETIME2 NOT NULL DEFAULT SYSUTCDATETIME()
);

CREATE INDEX idx_audit_log_actor ON audit_log (actor);
CREATE INDEX idx_audit_log_occurred_at ON audit_log (occurred_at);