import asyncio
import json
import uuid

import aiohttp

room :str = ''

async def listener(l: asyncio.Condition):
    async with aiohttp.ClientSession() as session:

        ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
        login_data = {
            "id": uuid.uuid4().hex,
            "method": "base.login",
            "params": [
                "xxxx@qq.com",
                "123456"
            ]
        }
        await ws.send_str(json.dumps(login_data))
        resp = await ws.receive()
        print(resp.data)
        async with l:
            await l.wait()
        data = {
            "id": "1",
            "method": "room.in",
            "params": [
                room
            ]
        }
        await ws.send_str(json.dumps(data))
        print('get in room: ', (await ws.receive()).data)
        await ws.send_str(json.dumps({'id': '', 'method': 'room.mates', 'params': [room]}))
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
        ws = await session.ws_connect('ws://127.0.0.1:8000/ws')
        login_data = {
            "id": uuid.uuid4().hex,
            "method": "base.login",
            "params": [
                "111@qq.com",
                "123456"
            ]
        }
        await ws.send_str(json.dumps(login_data))
        resp = await ws.receive()
        print(resp.data)

        await ws.send_str(json.dumps({'id': uuid.uuid4().hex, 'method': 'room.create', 'params': []}))
        resp = await ws.receive()
        global room
        room = json.loads(resp.data)['data']
        print('room: ', room)
        async with l:
            l.notify()
        await asyncio.sleep(20)


async def main():
    create_lock = asyncio.Condition()
    await asyncio.gather(listener(create_lock), sender(create_lock))


if __name__ == '__main__':
    asyncio.run(main())
