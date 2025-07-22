migrations/001_initial_schema.sql
This file is for reference if you prefer manual SQL migrations.
GORM AutoMigrate handles this for you in main.go

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP WITH TIME ZONE,    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    number_of_drone INTEGER NOT NULL,
    location VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    quantity INTEGER NOT NULL,
    status VARCHAR(50),
    category_id INTEGER NOT NULL,
    remarks TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id),
);
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
);

CREATE TABLE transaction_borrows (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    borrow_quantity INTEGER NOT NULL,
    borrow_date DATE NOT NULL,
    due_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);

CREATE TABLE transaction_returns (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    item_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    borrow_id INTEGER NOT NULL,
    return_quantity INTEGER NOT NULL,
    return_date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
    FOREIGN KEY (borrow_id) REFERENCES borrows(id)
);

CREATE TABLE damage_reports (
    id SERIAL PRIMARY KEY,
    item_id INTEGER NOT NULL,
    reporter_id INTEGER NOT NULL,
    project_id INTEGER NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'Pending', 
    broken_drone INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (item_id) REFERENCES items(id),
    FOREIGN KEY (reporter_id) REFERENCES users(id),
    FOREIGN KEY (project_id) REFERENCES projects(id)
);
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    actiooon VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    record_id INTEGER NOT NULL,
    old_value VARCHAR(50) ,
    new_value VARCHAR(50) , 
    ip_address VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),

);