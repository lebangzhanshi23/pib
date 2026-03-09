-- PIB Database Schema
-- Migration: 001_init.sql

-- Questions table
CREATE TABLE IF NOT EXISTS questions (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    answer TEXT,
    summary TEXT,
    ef REAL DEFAULT 2.5,
    interval INTEGER DEFAULT 0,
    next_review_at DATETIME,
    status TEXT DEFAULT 'draft' CHECK(status IN ('draft', 'active', 'archived')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Question-Tags relationship
CREATE TABLE IF NOT EXISTS question_tags (
    question_id TEXT NOT NULL,
    tag_id TEXT NOT NULL,
    PRIMARY KEY (question_id, tag_id),
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

-- Review logs
CREATE TABLE IF NOT EXISTS review_logs (
    id TEXT PRIMARY KEY,
    question_id TEXT NOT NULL,
    grade INTEGER NOT NULL CHECK(grade IN (0, 1, 2)),
    reviewed_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_questions_next_review ON questions(next_review_at);
CREATE INDEX IF NOT EXISTS idx_questions_status ON questions(status);
CREATE INDEX IF NOT EXISTS idx_review_logs_question ON review_logs(question_id);
