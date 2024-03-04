DROP TABLE IF EXISTS dialogues;
DROP TABLE IF EXISTS messages;

DROP INDEX CONCURRENTLY IF EXISTS messages_dialogue_id_created_at_ind;