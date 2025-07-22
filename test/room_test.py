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
        print('get mates: ', (await ws.receive()).data)
        await asyncio.sleep(5)
        # count = 0
        # async for msg in ws:
        #     count += 1
        #     if msg.data == "ping":
        #         await ws.send_str("pong")
        #         print("heartbeat")
        #         continue
        #     print("listen:", msg.data)
        #     # if count > 20:
        #     #     break

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
            'UserIdBlackList': [2]
        }]}))
        resp = await ws.receive()
        print(resp.data)
        global room
        room = json.loads(resp.data)['data']
        print('room: ', room)
        async with l:
            l.notify()
        for i in range(100):
            await ws.send_str(json.dumps({
                'id': '', 'method': 'room.message', 'params': [room, 'hello' + str(i)]
            }))
            await asyncio.sleep(2)


async def main():
    create_lock = asyncio.Condition()
    await asyncio.gather(listener(create_lock), sender(create_lock))


if __name__ == '__main__':
    asyncio.run(main())
