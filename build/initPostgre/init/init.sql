CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    fio TEXT,
    username TEXT UNIQUE,
    pass TEXT,
    role TEXT,
    age SMALLINT,
    specialty TEXT,
    price INTEGER,
    rating REAL DEFAULT 0,
    teachers UUID[],
    students UUID[],
    requests UUID[]
);
