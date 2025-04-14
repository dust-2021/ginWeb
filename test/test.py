import asyncio
import json

import aiohttp


async def login(session: aiohttp.ClientSession) -> str:
    user = "mole"
    password = "123456"
    resp = await session.post("http://127.0.0.1:8000/api/login",
                              data=json.dumps({"username": user, "password": password}))
    info = await resp.json()
    print(info)
    return info['data']



async def create_user():
    session = aiohttp.ClientSession()
    token = await login(session)

    session.headers["Token"] = token
    resp = await session.get('http://127.0.0.1:8000/sapi/system/user/createPieces?count=2', )
    print(await resp.text())
    await session.close()


if __name__ == '__main__':
    asyncio.run(create_user())
