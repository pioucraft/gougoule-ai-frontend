CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE TABLE ai_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    api_key TEXT NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE models (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    provider_id UUID REFERENCES ai_providers(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW()
);



CREATE TABLE memory_cells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO memory_cells (content) VALUES ('The user uses Gougoule AI Frontend');

-- Step 1: Add a generated tsvector column (if it doesn't already exist)
ALTER TABLE messages
ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
  to_tsvector('english', content)
) STORED;

-- Step 2: Create a GIN index on the generated column
CREATE INDEX IF NOT EXISTS idx_messages_search_vector
ON messages
USING GIN (search_vector);
