-- Commenters registering with a password must confirm their email address
-- before they can log in. Existing commenters are assumed to be confirmed.
ALTER TABLE commenters
  ADD confirmedEmail BOOLEAN NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS commenterConfirmHexes (
  confirmHex               TEXT          NOT NULL  UNIQUE  PRIMARY KEY      ,
  commenterHex             TEXT          NOT NULL                           ,
  sendDate                 TIMESTAMP     NOT NULL
);
