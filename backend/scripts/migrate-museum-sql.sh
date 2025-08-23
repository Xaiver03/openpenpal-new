#!/bin/bash

# Direct SQL Migration for Museum Tables
# 博物馆表直接SQL迁移脚本

echo "🏛️ Starting Museum Tables SQL Migration..."
echo "========================================"

# 设置颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# PostgreSQL connection details
DB_NAME="openpenpal"
DB_USER=$(whoami)

echo -e "${YELLOW}Creating museum extended tables...${NC}"

# SQL command to create missing museum tables
SQL_COMMANDS="
-- Museum Tags Table
CREATE TABLE IF NOT EXISTS museum_tags (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(50) NOT NULL,
    count INTEGER DEFAULT 0,
    trending BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Museum Interactions Table
CREATE TABLE IF NOT EXISTS museum_interactions (
    id VARCHAR(36) PRIMARY KEY,
    entry_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    type VARCHAR(20) NOT NULL, -- view, like, share, bookmark
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (entry_id) REFERENCES museum_entries(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(entry_id, user_id, type)
);

-- Museum Reactions Table
CREATE TABLE IF NOT EXISTS museum_reactions (
    id VARCHAR(36) PRIMARY KEY,
    entry_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    reaction_type VARCHAR(50) NOT NULL, -- emotion, reflection, memory, inspiration
    comment TEXT,
    is_anonymous BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (entry_id) REFERENCES museum_entries(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Museum Submissions Table
CREATE TABLE IF NOT EXISTS museum_submissions (
    id VARCHAR(36) PRIMARY KEY,
    letter_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(200) NOT NULL,
    author_name VARCHAR(100) NOT NULL,
    display_preference VARCHAR(20) DEFAULT 'anonymous', -- anonymous, nickname, realname
    tags TEXT[], -- PostgreSQL array
    category VARCHAR(50),
    submission_reason TEXT,
    status VARCHAR(20) DEFAULT 'pending', -- pending, approved, rejected, withdrawn
    review_notes TEXT,
    reviewed_by VARCHAR(36),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (letter_id) REFERENCES letters(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_museum_interactions_entry_id ON museum_interactions(entry_id);
CREATE INDEX IF NOT EXISTS idx_museum_interactions_user_id ON museum_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_reactions_entry_id ON museum_reactions(entry_id);
CREATE INDEX IF NOT EXISTS idx_museum_reactions_user_id ON museum_reactions(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_submissions_letter_id ON museum_submissions(letter_id);
CREATE INDEX IF NOT EXISTS idx_museum_submissions_user_id ON museum_submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_museum_submissions_status ON museum_submissions(status);
CREATE INDEX IF NOT EXISTS idx_museum_tags_category ON museum_tags(category);
CREATE INDEX IF NOT EXISTS idx_museum_tags_trending ON museum_tags(trending);
"

# Execute SQL
psql -U "$DB_USER" -d "$DB_NAME" -c "$SQL_COMMANDS" 2>/dev/null

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Museum tables created successfully${NC}"
else
    echo -e "${RED}❌ Failed to create museum tables${NC}"
    echo "You may need to run: createdb openpenpal"
    exit 1
fi

# Verify tables
echo -e "\n${YELLOW}Verifying museum tables...${NC}"
psql -U "$DB_USER" -d "$DB_NAME" -c "\dt museum_*" 2>/dev/null

echo -e "\n${GREEN}🎉 Museum SQL migration completed!${NC}"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Test museum API endpoints with ./scripts/test-museum-apis.sh"
echo "2. Continue with frontend development"