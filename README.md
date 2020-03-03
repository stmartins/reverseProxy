
## first step:
    source .env && go install reverseProxy.go

execute reverseProxy and 3 http-server in 4 different terminals:

```console
TTY1@stf:~$ reverseProxy

TTY2@stf:~$ http-server -p 1331

TTY3@stf:~$ http-server -p 1332

TTY4@stf:~$ http-server -p 1333
```

- execute insomnia.app
- get request to localhost:1330
- request body: JSON
        {
            "proxy_condition": "a"
        }

switch between "a", "b" and "c" to see each request in appropriate server
