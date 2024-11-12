import websocket

ws = websocket.WebSocket()
ws.connect("ws://127.0.0.1:8080/echo")
ws.send("Hello, Oasis!")
ws.close()
