USE motd;

CREATE TABLE IF NOT EXISTS message_of_the_day (
    id      INT AUTO_INCREMENT PRIMARY KEY,
    message VARCHAR(255) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

INSERT INTO message_of_the_day (id, message) VALUES (1, 'No plans');
