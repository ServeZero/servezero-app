#!/bin/sh

sqlite3 init.sqlite3  "CREATE TABLE system (id TEXT PRIMARY KEY, value TEXT);"
sqlite3 init.sqlite3  "CREATE TABLE user (id INTEGER PRIMARY KEY, account TEXT, password TEXT);"
sqlite3 init.sqlite3  "CREATE TABLE domain (id INTEGER PRIMARY KEY, name TEXT, app_type TEXT, dir_name TEXT, hash TEXT, db_name TEXT, db_user TEXT, db_password TEXT, created_dt TEXT);"
sqlite3 init.sqlite3  "INSERT INTO user (account, password) values ('admin@example.com', '\$2a\$10\$qc8rQ5f9NWL5FVKDUghr.ejS3sMiVT/.RFYwHVhfHiuSudSaBNnxa');" # admin
sqlite3 init.sqlite3  "SELECT * FROM user";
sqlite3 init.sqlite3  "INSERT INTO system (id, value) values ('db_version', '20220214');"
