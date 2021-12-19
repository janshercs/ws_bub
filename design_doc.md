# Wassup Bub
## A reddit/twitter clone complete with up/down votes

### Functional requirements (Backend)
1. Users able to post threads ✅
2. Users able to read threads ✅
3. Users able to up/down vote
4. Users able to comment/subcomment
5. Server to rank threads according to freshness/activity/popularity
6. User able to see live updates on the popularity of topics (websocket connection)

### Functional requirements (Frontend)
1. Render threads in floating bubbles
2. Bubble size corresponding to server ranking
3. User able to see live updates on the popularity of topics

### Server Logic
#### Routing
The server currently has 3 endpoints:
1. `/` - for healthcheck.
2. `thread` - for CRUD all threads (to be deprecated).
3. `ws` - websocket endpoint for sending/receiving threads.
#### Websocket
The server has a websocket (client) manager which adds and removes client connections from its register. 
The websocket manager also has a `broadcast` function to push the latest list of `Threads` to all connected sockets.

When a websocket connection is connected to the server, a `go routine`, `ProcessThreadFromClient` will be called on that connection to read messages sent from the client; when a close message is received, the connection will be removed by the manager from the register.

#### Channels and Workers
The server has 2 channels: `threadChannel` and `sendChannel` and 2 types of workers (`go routines`): `threadSaver`, `socketUpdater`.

When a thread is received from any of the connected clients, the thread will be sent to the `threadChannel` where a `threadSaver` worker will dequeue the thread, and save it. When successfully saved, the `threadSaver` worker will send a `true` signal to the `sendChannel` where the `socketUpdater` worker will dequeue the signal and send the updated `threads` to all connected clients.

The workers can be started by the server by `StartWorkers()` method.

### API Reference
Communication between the front and backend services are centered around the `Thread` object.
Sample `Thread` Object
```json
{
  "ID": 0,
  "Content": "Sample Message.",
  "User": "Awesome_user",
  "UpVotesCount": 0,
  "DownVotesCount": 0
}
```
---

`POST /thread`
```json
{
  "Content": "Sample Message.",
  "User": "Awesome_user"
}
```
Response: `Thread`
```json
{
  "ID": 0,
  "Content": "Sample Message.",
  "User": "Awesome_user",
  "UpVotesCount": 0,
  "DownVotesCount": 0
}
```
---
`sendMessage /ws`
```json
{
  "Content": "Sample Message.",
  "User": "Awesome_user"
}
```
Response: `[]Thread`
```json
[
  {
    "ID": 0,
    "Content": "Sample Message.",
    "User": "Awesome_user",
    "UpVotesCount": 0,
    "DownVotesCount": 0
  },
  {
    "ID": 1,
    "Content": "Another sample Message.",
    "User": "Awesome_user",
    "UpVotesCount": 0,
    "DownVotesCount": 0
  },
]
```