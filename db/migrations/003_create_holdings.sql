-- +goose Up
CREATE TABLE holdings (
    user_id    TEXT    NOT NULL,
    stock_code TEXT    NOT NULL,
    shares     INTEGER NOT NULL CHECK (shares > 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, stock_code)
);

-- +goose Down
DROP TABLE IF EXISTS holdings;
