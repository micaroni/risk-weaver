-- +goose Up

CREATE TABLE workloads (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    namespace TEXT NOT NULL,
    environment TEXT NOT NULL,
    owner TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE workloads;