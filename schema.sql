CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    created_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS pvzs (
    id UUID PRIMARY KEY,
    registration_date TIMESTAMP NOT NULL,
    city VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS receptions (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    pvz_id UUID NOT NULL REFERENCES pvzs(id),
    status VARCHAR(50) NOT NULL
);

CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    date_time TIMESTAMP NOT NULL,
    type VARCHAR(50) NOT NULL,
    reception_id UUID NOT NULL REFERENCES receptions(id)
); 