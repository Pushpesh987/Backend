ALTER DATABASE postgres SET timezone TO 'Asia/Kolkata';

CREATE TYPE difficulty_enum AS ENUM ('easy', 'medium', 'hard');
CREATE TYPE event_status AS ENUM ('Upcoming', 'Ongoing', 'Completed');
CREATE TYPE workshop_status AS ENUM ('Upcoming', 'Ongoing', 'Completed');
CREATE TYPE project_status AS ENUM ('Ongoing', 'Completed');


CREATE TABLE IF NOT EXISTS auth (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    last_sign_in_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    email TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS badges (
    badge_id SERIAL PRIMARY KEY,
    badge_name VARCHAR(255) NOT NULL,
    level INT NOT NULL,
    points_required INT NOT NULL,
    streak_required INT NOT NULL
);

CREATE TABLE IF NOT EXISTS colleges (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    college_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID DEFAULT uuid_generate_v4() PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS communities (
    id SERIAL PRIMARY KEY,                
    name VARCHAR(255) NOT NULL,           
    description TEXT,                      
    created_at TIMESTAMP DEFAULT NOW()    
);

CREATE TABLE IF NOT EXISTS community_members (
    id SERIAL PRIMARY KEY,                 
    user_id UUID NOT NULL,                 
    community_id INT NOT NULL,             
    joined_at TIMESTAMP DEFAULT NOW(),     
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
    UNIQUE (user_id, community_id)        
);

CREATE TABLE IF NOT EXISTS connections (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    connection_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS education_levels (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    level_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    theme VARCHAR(255),
    description TEXT,
    date TIMESTAMP NOT NULL,
    location VARCHAR(255),
    entry_fee DECIMAL(10, 2),
    prize_pool DECIMAL(10, 2),
    media VARCHAR(255),
    registration_deadline DATE,
    organizer_name VARCHAR(255),
    organizer_contact VARCHAR(50),
    tags VARCHAR(255),
    attendee_count INT,
    status event_status NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS fields_of_study (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    field_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS interests (
    interest_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    interest_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS likes (
    user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    PRIMARY KEY (user_id, post_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS locations (
    id uuid PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,                 
    community_id INT NOT NULL,             
    user_id UUID NOT NULL,                 
    message TEXT NOT NULL,                 
    created_at TIMESTAMP DEFAULT NOW(),    
    FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS points_streak (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_points INT DEFAULT 0,
    current_streak INT DEFAULT 0,
    highest_streak INT DEFAULT 0,
    last_attempted DATE 
);

CREATE TABLE IF NOT EXISTS post_tags (
    id SERIAL PRIMARY KEY,
    post_id UUID REFERENCES posts(id),
    tag_id INT REFERENCES tags(id),
    UNIQUE (post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    media_url TEXT,
    likes_count INT DEFAULT 0,
    comments_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    domain VARCHAR(255),
    start_date DATE NOT NULL,
    end_date DATE,
    location VARCHAR(255),
    media VARCHAR(255),
    tags VARCHAR(255),
    team_members TEXT,
    status project_status NOT NULL,
    sponsors VARCHAR(255),
    project_link VARCHAR(255),
    goals TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS questions (
    question_id SERIAL PRIMARY KEY,
    question_text TEXT NOT NULL,
    options JSONB, 
    correct_answer TEXT NOT NULL,
    difficulty difficulty_enum NOT NULL,
    points INT NOT NULL DEFAULT 0,
    multiplier FLOAT DEFAULT 1.0,
    question_type VARCHAR(10) NOT NULL DEFAULT 'daily',
    CHECK (question_type IN ('daily', 'bonus','skill')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS quiz_attempts (
    attempt_id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    question_id INT REFERENCES questions(question_id) ON DELETE CASCADE,
    is_correct BOOLEAN,
    attempted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS shares (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    from_user_id UUID NOT NULL,
    to_user_id UUID NOT NULL,
    post_id UUID NOT NULL,
    shared_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS skills (
    skill_id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
    skill_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tags (
    id SERIAL PRIMARY KEY,
    tag VARCHAR UNIQUE
);

CREATE TABLE IF NOT EXISTS user_badges (
    user_badge_id SERIAL PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    badge_id INT REFERENCES badges(badge_id) ON DELETE CASCADE,
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY NOT NULL REFERENCES auth(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    name TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    username TEXT NOT NULL UNIQUE,
    age SMALLINT NOT NULL DEFAULT 0,
    gender TEXT NOT NULL DEFAULT '',
    profile_pic_url TEXT NOT NULL,
    profile_pic_size INT NOT NULL DEFAULT 0,
    profile_pic_content_type VARCHAR NOT NULL,
    profile_pic_storage_path TEXT NOT NULL,
    dob DATE NOT NULL,
    phone TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    location_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'::uuid,
    education_level_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'::uuid,
    field_of_study_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'::uuid,
    college_name_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000'::uuid,
    auth_id UUID REFERENCES auth(id),
    for_first_time BOOLEAN DEFAULT TRUE
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES locations(id) ON DELETE SET NULL,
    CONSTRAINT fk_education_level FOREIGN KEY (education_level_id) REFERENCES education_levels(id) ON DELETE SET NULL,
    CONSTRAINT fk_field_of_study FOREIGN KEY (field_of_study_id) REFERENCES fields_of_study(id) ON DELETE SET NULL,
    CONSTRAINT fk_college_name FOREIGN KEY (college_name_id) REFERENCES colleges(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS workshops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NULL,
    date TIMESTAMP NOT NULL,
    location VARCHAR(255) NULL,
    media TEXT NULL,
    entry_fee DECIMAL(10, 2) NULL,  
    duration VARCHAR(255) NULL,
    instructor_info TEXT NULL,
    tags VARCHAR(255) NULL,
    participant_limit INT NULL,   
    status workshop_status NOT NULL,
    registration_link TEXT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE OR REPLACE FUNCTION check_and_update_badges()
RETURNS TRIGGER AS $$
DECLARE
    current_badge_id INTEGER;  
    new_badge_id INTEGER;     
BEGIN
    SELECT badge_id INTO current_badge_id
    FROM user_badges
    WHERE user_id = NEW.user_id;

    SELECT b.badge_id INTO new_badge_id
    FROM badges b
    WHERE 
        b.points_required <= NEW.total_points 
        AND b.streak_required <= NEW.highest_streak 
    ORDER BY b.badge_id DESC 
    LIMIT 1;

    IF new_badge_id IS NOT NULL AND (current_badge_id IS NULL OR new_badge_id > current_badge_id) THEN
        UPDATE user_badges
        SET 
            badge_id = new_badge_id,
            earned_at = NOW()
        WHERE user_id = NEW.user_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_check_badges
AFTER UPDATE ON points_streak
FOR EACH ROW
EXECUTE FUNCTION check_and_update_badges();

CREATE OR REPLACE FUNCTION update_points_and_streak()
RETURNS TRIGGER AS $$
DECLARE
    question_points INT; 
    question_multiplier FLOAT; 
    points_to_add INT; 
    last_answered_time TIMESTAMP; 
    current_user_streak INT;
BEGIN

    SELECT points, multiplier INTO question_points, question_multiplier
    FROM questions
    WHERE question_id = NEW.question_id;

    points_to_add := question_points * question_multiplier;

    SELECT MAX(attempted_at) INTO last_answered_time
    FROM quiz_attempts
    WHERE user_id = NEW.user_id;

    SELECT current_streak INTO current_user_streak
    FROM points_streak
    WHERE user_id = NEW.user_id;

    IF EXISTS (SELECT 1 FROM points_streak WHERE user_id = NEW.user_id) THEN

        IF last_answered_time IS NOT NULL AND EXTRACT(DAY FROM CURRENT_DATE - last_answered_time) > 1 THEN
            UPDATE points_streak
            SET 
                current_streak = 0,
                total_points = total_points + points_to_add,
                highest_streak = highest_streak  
            WHERE user_id = NEW.user_id;
        ELSIF last_answered_time IS NOT NULL AND EXTRACT(DAY FROM CURRENT_DATE - last_answered_time) = 1 THEN
            UPDATE points_streak
            SET 
                current_streak = current_user_streak + 1,  
                total_points = total_points + points_to_add,
                highest_streak = GREATEST(current_user_streak + 1, highest_streak) 
            WHERE user_id = NEW.user_id;

        ELSE
            UPDATE points_streak
            SET 
                total_points = total_points + points_to_add
            WHERE user_id = NEW.user_id;
        END IF;

    ELSE
        INSERT INTO points_streak (user_id, total_points, current_streak, highest_streak)
        VALUES (
            NEW.user_id, 
            points_to_add, 
            CASE WHEN NEW.is_correct THEN 1 ELSE 0 END, 
            CASE WHEN NEW.is_correct THEN 1 ELSE 0 END
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_points_and_streak
AFTER INSERT ON quiz_attempts
FOR EACH ROW
EXECUTE FUNCTION update_points_and_streak();

