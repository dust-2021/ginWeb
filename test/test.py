import asyncio
import json

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

    ws = await session.ws_connect('ws://127.0.0.1:8000/ws', headers={
        "Token": await login(session),})
    data = {
        "id": "1",
        "method": "hello",
        "params": [

        ]
    }
    await ws.send_str(json.dumps(data))
    data = await ws.receive(5)
    print(data.data)
    await ws.ping()

    await ws.close()
    await session.close()


if __name__ == '__main__':
    asyncio.run(main())
