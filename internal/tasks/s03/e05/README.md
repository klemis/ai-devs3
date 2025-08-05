# S03E05 - Connections Task (Neo4j Graph Database)

## Overview

The S03E05 task implements a graph-based solution to find the shortest path between two users (Rafał and Barbara) using Neo4j graph database. It retrieves user and connection data from a MySQL database via API and builds a graph representation to solve the shortest path problem.

## Task Description

This task involves:
1. **Data Retrieval**: Get users and connections from MySQL database (reusing S03E03 logic)
2. **Graph Setup**: Connect to Neo4j and populate with user data
3. **Path Finding**: Use Cypher queries to find shortest path between Rafał and Barbara
4. **Result Submission**: Submit the path as comma-separated string to centrala API

## Prerequisites

### Environment Variables

Make sure the following environment variables are set in your `.zshrc` or `.bashrc`:

```bash
# Required for all tasks
export AI_DEVS_API_KEY="your_api_key_here"
export OPENAI_API_KEY="your_openai_key_here"

# Required for Neo4j (S03E05)
export NEO4J_URI="bolt://localhost:7687"     # or your Neo4j URI
export NEO4J_USER="neo4j"                    # your Neo4j username
export NEO4J_PASSWORD="your_password"       # your Neo4j password
```

### Neo4j Database Setup

#### Option 1: Docker (Recommended)
```bash
# Start Neo4j container
docker run -d \
  --name neo4j \
  -p 7474:7474 -p 7687:7687 \
  -e NEO4J_AUTH=neo4j/your_password \
  neo4j:latest

# Access Neo4j Browser at http://localhost:7474
```

#### Option 2: Local Installation
1. Download Neo4j Desktop from https://neo4j.com/download/
2. Create a new project and database
3. Set password and start the database
4. Note the connection details (usually bolt://localhost:7687)

## Usage

### Basic Command
```bash
./ai-devs3 s03e05
```

### Expected Output
```
Starting S03E05 connections task (Neo4j graph database)
Neo4j connection verified successfully
Starting connections data retrieval and graph processing...
Retrieved 1000 users and 5000 connections from MySQL
Clearing Neo4j database...
Creating user nodes in Neo4j...
Created 1000 user nodes
Creating connection relationships in Neo4j...
Created 5000 relationship edges
Finding shortest path from Rafał to Barbara...
Shortest path found: [Rafał, User1, User2, Barbara] (length: 3 steps)

=== Connections Task Results ===
Shortest Path: [Rafał, User1, User2, Barbara]
Path String: Rafał,User1,User2,Barbara
Path Length: 3 steps

=== Processing Statistics ===
Users loaded from MySQL: 1000
Connections loaded from MySQL: 5000
Neo4j nodes created: 1000
Neo4j relationships created: 5000
Processing time: 15.42 seconds
================================
Connections task successful!
Response: Task completed successfully
```

## Technical Implementation

### Architecture

```
MySQL API → GraphData → Neo4j → Cypher Query → Shortest Path → Result Submission
```

### Key Components

1. **MySQL Data Retrieval** (reused from S03E03):
   - Queries: `SELECT id, username FROM users`
   - Queries: `SELECT user1_id, user2_id FROM connections`
   - Uses existing HTTP client and error handling

2. **Neo4j Graph Operations**:
   - Node creation: `CREATE (u:Person {userId: $mysql_id, username: $username})`
   - Relationship creation: `CREATE (u1)-[:KNOWS]->(u2)`
   - Shortest path: `MATCH path = shortestPath((start:Person {username: 'Rafał'})-[:KNOWS*]-(end:Person {username: 'Barbara'}))`

3. **Critical Implementation Details**:
   - **Node Property**: Uses `userId` instead of `id` to avoid Neo4j internal ID conflicts
   - **Relationships**: Creates unidirectional `KNOWS` relationships
   - **UTF-8 Support**: Handles Polish characters (Rafał) correctly
   - **Error Handling**: Comprehensive error handling at each step

### File Structure
```
internal/tasks/s03/e05/
├── command.go      # Cobra command definition
├── handler.go      # Main task orchestration
├── service.go      # Business logic implementation
├── models.go       # Data structures
└── README.md       # This file

internal/neo4j/
└── client.go       # Neo4j client wrapper
```

## Troubleshooting

### Common Issues

1. **Neo4j Connection Failed**:
   ```
   Error: failed to connect to Neo4j: connection refused
   ```
   - Check if Neo4j is running: `docker ps | grep neo4j`
   - Verify connection settings in environment variables
   - Test connection: `curl http://localhost:7474`

2. **Authentication Error**:
   ```
   Error: Neo4j authentication failed
   ```
   - Verify NEO4J_USER and NEO4J_PASSWORD
   - Reset password if needed in Neo4j Browser

3. **No Path Found**:
   ```
   Error: no path found between Rafał and Barbara
   ```
   - Check if both users exist in the database
   - Verify connection data integrity
   - Examine Neo4j graph in browser

4. **Database Query Issues**:
   ```
   Error: database error: unauthorized
   ```
   - Verify AI_DEVS_API_KEY is correct
   - Check if API endpoint is accessible

### Debugging Commands

```bash
# Check Neo4j status
docker logs neo4j

# Verify environment variables
env | grep NEO4J
env | grep AI_DEVS

# Test basic connectivity
curl -u neo4j:your_password http://localhost:7474/db/data/

# View logs
./ai-devs3 s03e05 2>&1 | tee s03e05.log
```

### Neo4j Browser Queries

Access http://localhost:7474 and run these queries to debug:

```cypher
// Count all nodes
MATCH (n) RETURN count(n)

// Count all relationships
MATCH ()-[r]->() RETURN count(r)

// Find Rafał
MATCH (n:Person {username: 'Rafał'}) RETURN n

// Find Barbara
MATCH (n:Person {username: 'Barbara'}) RETURN n

// Manual shortest path check
MATCH path = shortestPath((start:Person {username: 'Rafał'})-[:KNOWS*]-(end:Person {username: 'Barbara'}))
RETURN [node in nodes(path) | node.username] as path

// View sample connections
MATCH (n:Person)-[:KNOWS]->(m:Person) RETURN n.username, m.username LIMIT 10
```

## Performance Considerations

- **Graph Size**: Handles thousands of nodes and relationships efficiently
- **Memory Usage**: Neo4j loads graph into memory for fast queries
- **Connection Pooling**: Neo4j driver manages connection pooling automatically
- **Query Optimization**: Uses Neo4j's optimized shortest path algorithms

## Security Notes

- Neo4j credentials are handled securely via environment variables
- Database is cleared and repopulated for each run (ensures data consistency)
- API keys are never logged or exposed in output
- Connection strings use encrypted bolt protocol when available