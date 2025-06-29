-- migrations/001_initial_schema.sql
-- This file is for reference if you prefer manual SQL migrations.
-- GORM AutoMigrate handles this for you in main.go
--
-- CREATE TABLE users (
--     id SERIAL PRIMARY KEY,
--     username VARCHAR(255) UNIQUE NOT NULL,
--     password VARCHAR(255) NOT NULL,
--     role VARCHAR(50) DEFAULT 'user' NOT NULL,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP WITH TIME ZONE
-- );
--
-- CREATE TABLE projects (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) UNIQUE NOT NULL,
--     description TEXT,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP WITH TIME ZONE
-- );
--
-- CREATE TABLE items (
--     id SERIAL PRIMARY KEY,
--     name VARCHAR(255) UNIQUE NOT NULL,
--     description TEXT,
--     quantity INTEGER NOT NULL,
--     status VARCHAR(50),
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP WITH TIME ZONE
-- );
--
-- CREATE TABLE transactions (
--     id SERIAL PRIMARY KEY,
--     user_id INTEGER NOT NULL,
--     item_id INTEGER NOT NULL,
--     project_id INTEGER NOT NULL,
--     quantity INTEGER NOT NULL,
--     type VARCHAR(50) NOT NULL, -- 'borrow' or 'return'
--     status VARCHAR(50), -- 'Pending', 'Approved', 'Rejected'
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP WITH TIME ZONE,
--     FOREIGN KEY (user_id) REFERENCES users(id),
--     FOREIGN KEY (item_id) REFERENCES items(id),
--     FOREIGN KEY (project_id) REFERENCES projects(id)
-- );
--
-- CREATE TABLE damage_reports (
--     id SERIAL PRIMARY KEY,
--     item_id INTEGER NOT NULL,
--     reporter_id INTEGER NOT NULL,
--     description TEXT NOT NULL,
--     status VARCHAR(50) DEFAULT 'Pending', -- 'Pending', 'Resolved', 'Irreparable'
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
--     deleted_at TIMESTAMP WITH TIME ZONE,
--     FOREIGN KEY (item_id) REFERENCES items(id),
--     FOREIGN KEY (reporter_id) REFERENCES users(id)
-- );