# Porous

A text-mode interface for managing ssh tunnels.

![porous](_assets/demo.gif)


## Install

You will need `go` installed and `GOBIN` in your `PATH`.

    go get -u github.com/velais/porous

## Config

Porous relies on your ~/.ssh/config. It looks for hosts that have a LocalForward or RemoteForward.


```Git Config
Host *
    ServerAliveInterval 240
    ServerAliveCountMax 2

Host tunnel-1
    User lknope
    Hostname remote-server-1.com
    LocalForward 4001 localhost:4001
    
Host tunnel-2
    User lknope
    Hostname remote-server-2.com
    RemoteForward 4002 localhost:4002
```



## Key Bindings

| action  | key                            |
| ------- | ------------------------------ |
| up      | <kbd>k</kbd> / <kbd>↑</kbd>    |
| down    | <kbd>j</kbd> / <kbd>↓</kbd>    |
| open    | <kbd>o</kbd> / <kbd>enter</kbd>|
| close   | <kbd>x</kdb>                   |
| reload  | <kbd>r</kdb>                   |
| info    | <kbd>i</kdb>                   |
| exit    | <kbd>q</kdb>                   | 


## Todo...

- [x] Better error messages when ssh/ssh_config cannot be found
- [x] Info screen to display full config for a host
- [ ] Specify config location
- [ ] Support ssh config includes
- [ ] Handle ctrl-c to cancel password prompt
- [ ] Show "loading" screen when waiting on ssh to finish
- [ ] Customize ssh command used?
- [ ] Ordering - state/name/host
