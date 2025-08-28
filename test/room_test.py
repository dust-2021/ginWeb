import asyncio
import json
import uuid

import aiohttp

room :str = ''

async def login(session: aiohttp.ClientSession, username, password) -> str:
    resp = await session.post("http://127.0.0.1:8000/api/login",
                              data=json.dumps({"username": username, "password": password}))
    info = await resp.json()
    return info['data']

async def listener(l: asyncio.Condition):
    async with aiohttp.ClientSession() as session:
        token = await login(session,"test", "123456")
        ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
        await ws.send_str(json.dumps({
            'id': '', 'method': 'base.auth', 'params': [token]
        }))
        print(await ws.receive())
        async with l:
            await l.wait()
        data = {
            "id": "get in room",
            "method": "room.in",
            "params": [
                room,
                "123456"
            ]
        }
        await ws.send_str(json.dumps(data))
        print('get in room: ', (await ws.receive()).data)
        await ws.send_str(json.dumps({'id': 'get mates', 'method': 'room.roommate', 'params': [room]}))

        async for msg in ws:
            print(msg.data)

async def sender(l: asyncio.Condition):
    async with aiohttp.ClientSession() as session:
        token = await login(session, "ddd", "ez2ymp")
        ws = await session.ws_connect('http://127.0.0.1:8000/ws')
        await ws.send_str(json.dumps({
            'id': '', 'method': 'base.auth', 'params': [token]
        }))
        print(await ws.receive())
        await ws.send_str(json.dumps({'id': uuid.uuid4().hex, 'method': 'room.create', 'params': [{
            'title': '宝宝巴士', 'maxMember': 16, 'password': '123456',
            'UserIdBlackList': [2], 'autoClose': False
        }]}))
        resp = await ws.receive()
        print(resp.data)
        global room
        room = json.loads(resp.data)['data']
        # for i in range(20):
        #     await ws.send_str(json.dumps({'id': uuid.uuid4().hex, 'method': 'room.create', 'params': [{
        #         'title': f'宝宝巴士{i}', 'maxMember': 16, 'password': '123456',
        #         'UserIdBlackList': [2]
        #     }]}))
        print('room: ', room)
        async with l:
            l.notify()
        for i in range(100):
            # await ws.send_str(json.dumps({
            #     'id': '', 'method': 'room.message', 'params': [room, 'hello' + str(i)]
            # }))
            await asyncio.sleep(2)


async def main():
    create_lock = asyncio.Condition()
    await asyncio.gather(listener(create_lock), sender(create_lock))


if __name__ == '__main__':
    asyncio.run(main())
