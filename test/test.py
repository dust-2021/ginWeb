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


async def ws_test():
    session = aiohttp.ClientSession()

    ws = await session.ws_connect('ws://127.0.0.1:8000/ws?channel=hall')
    data = {
        "id": "0",
        "method": "login",
        "params": [
            "13900004990",
            "123456"
        ]
    }
    # await ws.send_str(json.dumps(data))
    # data = await ws.receive(20)
    # print(data.data)
    data = {
        "id": "1",
        "method": "subscribe",
        "params": [
            "hello",
            # "time"
        ]
    }
    await ws.send_str(json.dumps(data))
    count = 0
    async for msg in ws:
        count += 1
        print(msg.data)
        if count > 20:
            break
    await ws.close()
    await session.close()


async def ws_test2():
    session = aiohttp.ClientSession()
    ws = await session.ws_connect('ws://127.0.0.1:8000/ws?channel=hall')
    for i in range(10):
        await ws.send_str(json.dumps({
            "id": str(i),
            "method": "broadcast",
            "params": [
                "hello everyone"
            ]
        }))
    await ws.close()
    await session.close()


if __name__ == '__main__':
    loop = asyncio.get_event_loop()
    loop.run_until_complete(asyncio.gather(ws_test(), ws_test2()))
