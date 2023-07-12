ALTER TABLE user_session ADD COLUMN archived_at date DEFAULT now();
ALTER TABLE user_session ADD COLUMN created_at date DEFAULT now();