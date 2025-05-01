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
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

INSERT INTO memory_cells (content, name) VALUES ('The user uses Gougoule AI Frontend', 'general');

