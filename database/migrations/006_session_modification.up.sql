ALTER TABLE user_session DROP COLUMN archived_at;
ALTER TABLE user_session DROP COLUMN created_at ;

ALTER TABLE user_session ADD COLUMN archived_at date DEFAULT NULL;
ALTER TABLE user_session ADD COLUMN created_at date DEFAULT now();