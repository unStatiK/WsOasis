import websocket

ws = websocket.WebSocket()
ws.connect("ws://127.0.0.1:8080/oasis")
ws.send('{"oasis_id" : "feed1", "message" : "Hello, Oasis!"}')
ws.close()
