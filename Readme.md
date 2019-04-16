## wait-host

`wait-host` is an API and a command line application that will wait on the availability of a host and TCP port.  
It is useful for synchronizing interdependent services, such as linked docker containers. 
You can use it to wait for a database to be ready, for php-fpm connection, ...


## Command line

```
Usage:
  wait-host [flags]

Examples:
    wait-host mysql:3306       	Wait indefinitely for port 3306 to be available on host mysql
    wait-host http://google.com Wait indefinitely for port 80 to be available on host google.com
    wait-host mysql:3306 -t 15  Wait a maximum of 15s for port 3306 to be available on host mysql
	

Flags:
  -d, --delay uint       Delay in seconds, before trying to contact the host.
  -h, --help             help for wait-host
  -H, --host string      Host or IP under test.
  -m, --message string   Delay message. (default "Waiting for connection on %s:%d")
  -p, --port uint        TCP port under test.
  -q, --quiet            Don't output any status messages.
  -t, --timeout uint     Timeout in seconds, zero for no timeout.
```

### Error Codes

The following error codes are returned:

|      |    |
|------|----|
| `0`  | The specified port on the host is accepting connections. |
| `1`  | A timeout occured waiting for the port to open. |
| `2`  | Un unknown error occured waiting for the port to open. The program cannot establish whether the port is open or not. |

## API

```
import "github.com/go-mods/wait-host"

func main() {
	wh, _ := waithost.New("mysql:3306")
	wh.Timeout = time.Duration(30) * time.Second
	wh.ConnectTimeout = time.Duration(1) * time.Second
	wh.SetWaitMessage("Connecting to {host}:{port}")
	wh.SetRetryMessage("Trying to connect to {host}:{port}")
	wh.SetSuccessMessage("successfully connected to {host}:{port}")
	wh.SetTimeoutMessage("Timeout trying to connect to {host}:{port}")
	_ = wh.Wait()
}
```
