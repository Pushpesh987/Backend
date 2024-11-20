
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name TEXT, 
    last_name TEXT, 
    username TEXT, 
    email TEXT NOT NULL UNIQUE, 
    password TEXT NOT NULL,
    gender TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, 
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, 
    profile_photo_url TEXT 
);
