CREATE TABLE if NOT EXISTS tasks (
    id VARCHAR(100) PRIMARY KEY,
    priority INT,
    taskType VARCHAR(30),
    status VARCHAR(60),
    backOffDuration BIGINT,
    payload JSONB,
    createdAt TIMESTAMP,
    createdBy VARCHAR(30),
    startedAt  TIMESTAMP,
    finishedAt  TIMESTAMP,
    error VARCHAR(100)
)