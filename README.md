# Project ika

This project is based on a different project that I made as a job application in the past.

I wanted to review it since some years have passed and I evolved, I hope, and actually have experience making a chat application for work with PHP.

Since I am learning `golang` it makes sense to me to make the challenge with go.

## Objective
To write a very simple ‘chat’ application backend in GO. 
Users should be able to send text messages to each other.
Users should be able to get messages sent to them selves together with the author information. 

The users and messages should be stored in a database. 
All communication between the client and server should happen over a simple JSON based protocol over HTTP. 

## Why name the project Ika? 

Ika meand squid in Japanese, but more important it is a simple mutation to ICA, which stands for Internet Chat Application. Which is in itself a play on the name of the IRC Protocol.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Run the application
```bash
make run
```
Create DB container
```bash
make docker-run
```

Shutdown DB Container
```bash
make docker-down
```

DB Integrations Test:
```bash
make itest
```

Live reload the application:
```bash
make watch
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```
