create table IF NOT EXISTS "public"."todo" (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    done bool NOT NULL,
    created_at TIMESTAMP NOT NULL
);