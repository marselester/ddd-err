package pg

// Schema is db schema which must be created before working with UserService.
const Schema = `
CREATE TABLE IF NOT EXISTS account (
    id varchar(27),
    username varchar(40) NOT NULL,
    PRIMARY KEY(id),
    UNIQUE(username)
);
`
