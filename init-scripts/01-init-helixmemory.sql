-- Initialize HelixMemory databases
-- Creates databases for Cognee, Mem0, and Letta

-- Cognee database
CREATE DATABASE cognee;
CREATE USER cognee WITH ENCRYPTED PASSWORD 'cognee';
GRANT ALL PRIVILEGES ON DATABASE cognee TO cognee;

-- Mem0 database
CREATE DATABASE mem0;
CREATE USER mem0 WITH ENCRYPTED PASSWORD 'mem0';
GRANT ALL PRIVILEGES ON DATABASE mem0 TO mem0;

-- Letta database
CREATE DATABASE letta;
CREATE USER letta WITH ENCRYPTED PASSWORD 'letta';
GRANT ALL PRIVILEGES ON DATABASE letta TO letta;

-- Enable pgvector extension for all databases
\c cognee
CREATE EXTENSION IF NOT EXISTS vector;

\c mem0
CREATE EXTENSION IF NOT EXISTS vector;

\c letta
CREATE EXTENSION IF NOT EXISTS vector;

\c helixmemory
CREATE EXTENSION IF NOT EXISTS vector;
