CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT UNIQUE,
  password_hash TEXT,
  created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS urls (
  id TEXT PRIMARY KEY,           -- the short code
  original_url TEXT NOT NULL,
  user_id UUID REFERENCES users(id),
  created_at TIMESTAMP DEFAULT now()
);
