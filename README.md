# TCP-chat

## How to run:

```
git clone https://github.com/Vlad1slavZhuk/tcp-chat.git

cd tcp-chat

go run server.go
```

Open another terminal and write:

```
go run client.go
```

## Commands on tcp-chat:

`/help` - lists all commands.

`/list` - lists all chat room.

`/create foo` - creates a chat room with name `foo`.

`/del foo` - deletes a chat room.

`/join foo` - joins a chat room named foo.

`/leave` - leaves the current chat room.

`/name foo` - changes your name to foo.

`/quit` - quits the program.