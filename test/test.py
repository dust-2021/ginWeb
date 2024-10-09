import asyncio
import json

import aiohttp


async def main():
    user = "13900004990"
    password = "123456"

    session = aiohttp.ClientSession()
    resp = await session.post("http://127.0.0.1:8000/api/login",
                              data=json.dumps({"username": user, "password": password}))
    info = await resp.json()
    token = info['data']
    ws = await session.ws_connect('ws://127.0.0.1:8000/ws', headers={
        "Token": token})
    data = {
        "id": "1",
        "method": "test.get",
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
