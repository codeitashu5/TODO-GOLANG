CREATE TABLE user_session(
    session_token TEXT NOT NULL,
    user_id INT REFERENCES users(id) NOT NULL
);



