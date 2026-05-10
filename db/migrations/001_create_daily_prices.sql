-- +goose Up
CREATE TABLE daily_prices (
    code        TEXT             NOT NULL,
    date        DATE             NOT NULL,
    name        TEXT             NOT NULL DEFAULT '',
    industry    TEXT             NOT NULL DEFAULT '',
    sector      TEXT             NOT NULL DEFAULT '',
    open        DOUBLE PRECISION,
    high        DOUBLE PRECISION,
    low         DOUBLE PRECISION,
    close       DOUBLE PRECISION,
    volume      DOUBLE PRECISION,
    saved_at    TIMESTAMPTZ      NOT NULL DEFAULT now(),
    PRIMARY KEY (code, date)
);

CREATE INDEX idx_daily_prices_code_date_desc ON daily_prices (code, date DESC);

-- +goose Down
DROP TABLE IF EXISTS daily_prices;
