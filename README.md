# Wassup Bub! 💬
> It's called wassup bub, cus I call people "bub", some times... and also, chat bubbles.
## Synopsis 📖

An example backend server that handles 
* Routing
* Websocket connections
* Storage (in flatfiles; json)
  
It currently works like a Chat application where the frontend connects via a websocket on the `/ws` path.

## Motivation 💪

This is a practice project to work on spinning up a server in `Go` and also for working with concurrency and websockets.

## Installation 📀

Latest binary of the server can be built in the `./cmd/server` folder by running `go build`. The resulting binary can be run on *port 5000* (default) using `./server` (or whatever the binary name is)

## Code Example 🤓

A live example can be found [here](https://wassup-bub.netlify.app/).
Code to the frontend can be found in this [repository](https://github.com/janshercs/ws_bub_frontend).
For in depth mechanics on server logic, see the [design document](./design_doc.md).

## File Structure 📁

Build/deploy folder: `/cmd/server`
Data related code: `/flat_FS.go` 
Server related code: `/server.go`
Thread related code: `/thread.go`
Websocket related code: `/client_websocket.go`

## Tests 👍

Tests for this repo are in `*_test.go` files and can be run with the command `go test` from the root folder.

## Contributors 🙆🙆‍♀️

I have not set up the repository for contributors (not expecting any), but should you be interested for some reason, feel free to message me on Github!
Should you have suggestions or critiques, feel free to open an `Discussion` and I'll take it under advisement!

## License 💸

Errr... it's FOSS? 