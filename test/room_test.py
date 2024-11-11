import asyncio
import json
import uuid

import aiohttp


async def listener(l: asyncio.Condition):
    session = aiohttp.ClientSession()

    ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
    login_data = {
        "id": uuid.uuid4().hex,
        "method": "login",
        "params": [
            "xxxx@qq.com",
            "123456"
        ]
    }
    await ws.send_str(json.dumps(login_data))
    print(await ws.receive())
    await l.wait()
    data = {
        "id": "1",
        "method": "room.in",
        "params": [
            # "hello",
            # "hall"
        ]
    }
    await ws.send_str(json.dumps(data))
    print(await ws.receive())
    count = 0
    async for msg in ws:
        count += 1
        if msg.data == "ping":
            await ws.send_str("pong")
            print("heartbeat")
            continue
        print("listen:", msg.data)
        # if count > 20:
        #     break
    await ws.close()
    await session.close()

async def sender(l: asyncio.Condition):
    session = aiohttp.ClientSession()
    ws = await session.ws_connect('ws://127.0.0.1:8000/ws')



async def main():
    create_lock = asyncio.Condition()

    pass


if __name__ == '__main__':
    asyncio.run(main())
