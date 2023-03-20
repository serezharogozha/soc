CREATE TABLE IF NOT EXISTS public.users(
    id serial PRIMARY KEY,
    first_name VARCHAR (50) NOT NULL,
    last_name VARCHAR (50) NOT NULL,
    age INTEGER NOT NULL,
    biography VARCHAR (255) NOT NULL,
    city VARCHAR (50) NOT NULL,
    password VARCHAR (50) NOT NULL
);