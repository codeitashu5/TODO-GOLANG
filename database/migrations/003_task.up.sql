CREATE TABLE task(
    id serial  PRIMARY KEY,
    task TEXT NOT NULL,
    is_task_completed BOOLEAN,
    user_id INT REFERENCES users(id) NOT NULL
);