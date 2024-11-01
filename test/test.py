import asyncio
import json
import time
import datetime

import aiohttp


async def login(session: aiohttp.ClientSession) -> str:
    user = "13900004990"
    password = "123456"
    resp = await session.post("http://127.0.0.1:8000/api/login",
                              data=json.dumps({"username": user, "password": password}))
    info = await resp.json()
    return info['data']


async def listener():
    session = aiohttp.ClientSession()

    ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
    data = {
        "id": "1",
        "method": "channel.subscribe",
        "params": [
            # "hello",
            "hall"
        ]
    }
    await ws.send_str(json.dumps(data))
    await ws.receive()
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


async def sender():
    session = aiohttp.ClientSession()
    ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
    login_data = {
        "id": "0",
        "method": "base.login",
        "params": [
            "xxxx@qq.com",
            "123456"
        ]
    }
    await ws.send_str(json.dumps(login_data))
    await ws.receive()
    await ws.send_str(json.dumps({
        "id": "0",
        "method": "channel.subscribe",
        "params": [
            "hall",
        ]
    }))
    await ws.receive()
    for i in range(10):
        await ws.send_str(json.dumps({
            "id": str(i),
            "method": "channel.broadcast",
            "params": [
                "hall",
                "hello everyone"
            ]
        }))
    async for resp in ws:
        if resp.data == "ping":
            print("heartbeat")
            await ws.send_str("pong")
            continue
        print("send result:", resp.data)
    await ws.close()
    await session.close()


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(asyncio.gather(listener(), sender()))
