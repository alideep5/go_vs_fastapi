### **Running PostgreSQL with Docker Compose**

First, run Docker Compose to start PostgreSQL:

```sh
docker-compose up -d
```

### **PostgreSQL Schema Creation Script**

```sql
CREATE TABLE posts (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    likes INT NOT NULL DEFAULT 0,
    comments INT NOT NULL DEFAULT 0,
    shares INT NOT NULL DEFAULT 0
);

```

### **Seed Data Generation Script**

```sql
DO $$
BEGIN
    FOR i IN 1..100 LOOP
        INSERT INTO posts (user_id, content, created_at, likes, comments, shares)
        VALUES (
            (1 + (random() * 10)::INT), -- Random user_id between 1 and 10
            'This is a sample post content #' || i, -- Unique content
            NOW() - (interval '1 hour' * (random() * 720)::INT), -- Random timestamp within the last 30 days
            (random() * 100)::INT, -- Random likes between 0 and 100
            (random() * 50)::INT, -- Random comments between 0 and 50
            (random() * 20)::INT -- Random shares between 0 and 20
        );
    END LOOP;
END $$;

```