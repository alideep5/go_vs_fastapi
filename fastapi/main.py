from fastapi import FastAPI
import asyncio

app = FastAPI()


async def fake_io_operation():
    await asyncio.sleep(1)
    return "I/O operation complete!"


@app.get("/")
async def read_root():
    return {"message": "Hello, World!"}


@app.get("/task")
async def run_async_task():
    result = await fake_io_operation()
    return {"message": result}
