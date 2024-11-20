ALTER DATABASE postgres SET timezone TO 'Asia/Kolkata';


CREATE TABLE IF NOT EXISTS auth (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_sign_in_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    username TEXT NOT NULL,
    password TEXT NOT NULL,
    UNIQUE (username)
);


CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL REFERENCES auth(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    age SMALLINT NOT NULL DEFAULT 0,
    gender TEXT NOT NULL DEFAULT '',
    profile_pic_url TEXT NOT NULL,
    profile_pic_size INT NOT NULL DEFAULT 0,
    profile_pic_content_type VARCHAR NOT NULL,
    profile_pic_storage_path TEXT NOT NULL,
    dob DATE NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    location_id UUID NOT NULL,
    education_level_id UUID NOT NULL,
    field_of_study_id UUID NOT NULL,
    college_name_id UUID NOT NULL,
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL,
    CONSTRAINT fk_education_level FOREIGN KEY (education_level_id) REFERENCES education_levels(id) ON DELETE SET NULL,
    CONSTRAINT fk_field_of_study FOREIGN KEY (field_of_study_id) REFERENCES fields_of_study(id) ON DELETE SET NULL,
    CONSTRAINT fk_college_name FOREIGN KEY (college_name_id) REFERENCES colleges(id) ON DELETE SET NULL
);



CREATE TABLE IF NOT EXISTS locations (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS education_levels (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    level_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS fields_of_study (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    field_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS colleges (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    college_name TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS interests (
    interest_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    interest_name TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS skills (
    skill_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    skill_name TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS user_interests (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id UUID NOT NULL REFERENCES interests(interest_id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, interest_id)
);


CREATE TABLE IF NOT EXISTS user_skills (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id UUID NOT NULL REFERENCES skills(skill_id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, skill_id)
);

