CREATE TABLE IF NOT EXISTS public.messages(
                                            dialogue_id int NOT NULL,
                                            from_user_id int NOT NULL,
                                            to_user_id int NOT NULL,
                                            text text NOT NULL,
                                            created_at timestamp NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.dialogues(
                                        id serial PRIMARY KEY,
                                        from_user_id int NOT NULL,
                                        to_user_id int NOT NULL,
                                        created_at timestamp NOT NULL DEFAULT now()
);

CREATE INDEX CONCURRENTLY messages_dialogue_id_created_at_ind ON messages (dialogue_id, created_at);