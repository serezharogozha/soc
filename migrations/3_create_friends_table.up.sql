CREATE TABLE IF NOT EXISTS public.friends(
    id serial PRIMARY KEY,
    user_id int NOT NULL,
    friend_id int NOT NULL
);