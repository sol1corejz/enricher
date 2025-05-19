CREATE TABLE users (
                       id BIGSERIAL PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       surname VARCHAR(255) NOT NULL,
                       patronymic VARCHAR(255),
                       age INTEGER NOT NULL,
                       sex VARCHAR(10) NOT NULL CHECK (sex IN ('male', 'female')),
                       country JSONB NOT NULL
);