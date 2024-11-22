from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import asyncpg
from datetime import datetime, timedelta
from typing import List

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


async def calculate_score(post: Post) -> float:
    time_decay = (
        datetime.utcnow() - post.created_at
    ).total_seconds() / 3600  # Decay in hours
    return (2 * post.likes) + (3 * post.comments) + (5 * post.shares) - time_decay


@app.on_event("startup")
async def startup():
    """Initialize the database connection."""
    app.state.db = await asyncpg.create_pool(DATABASE_URL)


@app.on_event("shutdown")
async def shutdown():
    """Close the database connection."""
    await app.state.db.close()


@app.get("/top-posts", response_model=List[Post])
async def get_top_posts():
    """Fetch top 10 posts based on engagement score."""
    query = """
        SELECT id, user_id, content, created_at, likes, comments, shares
        FROM posts
        LIMIT 100
    """
    try:
        async with app.state.db.acquire() as conn:
            rows = await conn.fetch(query)
    except Exception as e:
        raise HTTPException(status_code=500, detail="Database error: " + str(e))

    if not rows:
        raise HTTPException(status_code=404, detail="No posts found")

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
        post.engagement_score = await calculate_score(post)
        posts.append(post)

    # Sort by engagement score in descending order
    posts = sorted(posts, key=lambda p: p.engagement_score, reverse=True)

    # Return the top 10 posts
    return posts[:10]
