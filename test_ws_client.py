import asyncio
import websockets
import json

WS_URL = "ws://indicated-method-managed-omissions.trycloudflare.com/ws"  # Change if your backend runs on a different port

async def main():
    async with websockets.connect(WS_URL) as ws:
        print(f"Connected to {WS_URL}")
        while True:
            msg = await ws.recv()
            try:
                data = json.loads(msg)
                print(json.dumps(data, indent=2))
            except Exception as e:
                print(f"Error parsing message: {e}")
                print(msg)

if __name__ == "__main__":
    asyncio.run(main())
