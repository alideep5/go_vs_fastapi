from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from datetime import datetime, timedelta
from typing import List
import asyncpg
import random

app = FastAPI()

DATABASE_URL = "postgresql://order:orderpassword@localhost:5432/order"


class Post(BaseModel):
    id: int
    user_id: int
    content: str
    created_at: datetime
    likes: int
    comments: int
    shares: int
    engagement_score: float = None


class CreatePostRequest(BaseModel):
    content: str


@app.on_event("startup")
async def startup():
    app.state.db_pool = await asyncpg.create_pool(DATABASE_URL)


@app.on_event("shutdown")
async def shutdown():
    await app.state.db_pool.close()


@app.post("/create-and-fetch", response_model=List[Post])
async def create_and_fetch_top_posts(request: CreatePostRequest):
    # Fetch the last user ID
    fetch_last_user_query = "SELECT user_id FROM posts ORDER BY id DESC LIMIT 1"
    try:
        async with app.state.db_pool.acquire() as conn:
            last_user_id = await conn.fetchval(fetch_last_user_query)
            if last_user_id is None:
                last_user_id = 0
            user_id = last_user_id + 1
    except Exception as e:
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    # Insert the new post
    content = request.content
    created_at = datetime.utcnow() - timedelta(hours=random.randint(0, 720))
    likes = random.randint(0, 100)
    comments = random.randint(0, 50)
    shares = random.randint(0, 20)

    insert_post_query = """
        INSERT INTO posts (user_id, content, created_at, likes, comments, shares)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    """
    try:
        async with app.state.db_pool.acquire() as conn:
            post_id = await conn.fetchval(
                insert_post_query, user_id, content, created_at, likes, comments, shares
            )
    except Exception as e:
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    # Fetch the top 100,000 posts
    fetch_posts_query = """
        SELECT id, user_id, content, created_at, likes, comments, shares
        FROM posts
        ORDER BY created_at DESC
        LIMIT 25000
    """
    try:
        async with app.state.db_pool.acquire() as conn:
            rows = await conn.fetch(fetch_posts_query)
    except Exception as e:
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    # Calculate engagement scores and sort
    posts = []
    for row in rows:
        post = Post(
            id=row["id"],
            user_id=row["user_id"],
            content=row["content"],
            created_at=row["created_at"],
            likes=row["likes"],
            comments=row["comments"],
            shares=row["shares"],
        )
        post.engagement_score = (
            (2 * post.likes)
            + (3 * post.comments)
            + (5 * post.shares)
            - ((datetime.utcnow() - post.created_at).total_seconds() / 3600)
        )
        posts.append(post)

    posts.sort(key=lambda p: p.engagement_score, reverse=True)
    return posts[:10]
