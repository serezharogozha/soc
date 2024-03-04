CREATE TABLE IF NOT EXISTS public.posts(
    id serial PRIMARY KEY,
    text VARCHAR (255) NOT NULL,
    user_id int NOT NULL
);