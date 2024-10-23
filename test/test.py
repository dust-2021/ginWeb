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



async def main():


    session = aiohttp.ClientSession()

    ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
    data = {
        "id": "0",
        "method": "login",
        "params": [
            "13900004990",
            "123456"
        ]
    }
    await ws.send_str(json.dumps(data))
    data = await ws.receive(20)
    print(data.data)
    # await ws.ping('ping'.encode())
    # resp = await ws.receive(10)
    # print(resp.data)
    # await asyncio.sleep(62)

    data = {
        "id": "1",
        "method": "time",
        "params": []
    }
    await ws.send_str(json.dumps(data))
    resp = await ws.receive()
    info = json.loads(resp.data)
    print(datetime.datetime.fromtimestamp(info['data'] / 1000))
    await ws.close()
    await session.close()


if __name__ == '__main__':
    asyncio.run(main())
