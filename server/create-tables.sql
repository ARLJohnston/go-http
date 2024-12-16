DROP TABLE IF EXISTS album;
CREATE TABLE album (
  id         INT AUTO_INCREMENT NOT NULL,
  title      VARCHAR(128) NOT NULL,
  artist     VARCHAR(255) NOT NULL,
	score      INT NOT NULL,
  cover      VARCHAR(512) DEFAULT NULL,
  PRIMARY KEY (`id`)
);

INSERT INTO album
  (title, artist, price, cover)
VALUES
  ('Blue Train', 'John Coltrane', 20, 'https://upload.wikimedia.org/wikipedia/en/thumb/6/68/John_Coltrane_-_Blue_Train.jpg/220px-John_Coltrane_-_Blue_Train.jpg'),
	('Future Nostalgia', 'Dua Lipa', 56, 'https://upload.wikimedia.org/wikipedia/en/thumb/f/f5/Dua_Lipa_-_Future_Nostalgia_%28Official_Album_Cover%29.png/220px-Dua_Lipa_-_Future_Nostalgia_%28Official_Album_Cover%29.png')
  ('Giant Steps', 'John Coltrane', 10, 'https://upload.wikimedia.org/wikipedia/en/thumb/2/2a/Coltrane_Giant_Steps.jpg/220px-Coltrane_Giant_Steps.jpg'),
  ('Jeru', 'Gerry Mulligan', 17, 'https://upload.wikimedia.org/wikipedia/en/thumb/c/ca/Jeru_%28album%29.jpg/220px-Jeru_%28album%29.jpg'),
  ('Sarah Vaughan', 'Sarah Vaughan', 34, 'https://upload.wikimedia.org/wikipedia/en/thumb/d/d8/Sarah_Vaughan.jpg/220px-Sarah_Vaughan.jpg');
