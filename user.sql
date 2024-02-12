CREATE DATABASE user_cache_dev;

CREATE TABLE users (
  id         INT AUTO_INCREMENT NOT NULL,
  name      VARCHAR(128) NOT NULL,
  email     VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO users
  (name, email)
VALUES
  ('John Good', 'johngood@eee.com'),
  ('John Bad', 'johngood111@eee.com'),
  ('John Doe', 'johngood222@eee.com'),
  ('John Kowalski','johngood333@eee.com');