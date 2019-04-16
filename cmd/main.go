package main

import (
	"fmt"
	"github.com/go-mods/wait-host"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var wh *waithost.WaitHost = nil

var cmd = &cobra.Command{
	Use:   "wait-host",
	Short: "Wait for host and port availability",
	Long: `
wait-host is useful for synchronizing interdependent services, 
such as linked docker containers. You can use it to wait for 
a database to be ready, for php-fpm connection, ...`,
	Example: `    wait-host mysql:3306       	Wait indefinitely for port 3306 to be available on host mysql
    wait-host http://google.com Wait indefinitely for port 80 to be available on host google.com
    wait-host mysql:3306 -t 15  Wait a maximum of 15s for port 3306 to be available on host mysql
	`,

	Run: func(cmd *cobra.Command, args []string) {

		// host and scheme
		var host string
		if len(args) > 0 {
			host = args[0]
		} else {
			host, _ = cmd.Flags().GetString("host")
		}
		if len(host) > 0 {
			if w, err := waithost.New(host); err != nil {
				os.Exit(2)
			} else {
				wh = w
			}
		}

		//
		if wh == nil {
			os.Exit(2)
		}

		// Port
		port, _ := cmd.Flags().GetUint("port")
		if port > 0 {
			wh.Port = port
		}

		// Delay
		delay, _ := cmd.Flags().GetUint("delay")
		if delay > 0 {
			wh.ConnectTimeout = time.Duration(delay) * time.Second
		}

		// Timeout
		timeout, _ := cmd.Flags().GetUint("timeout")
		if timeout > 0 {
			wh.Timeout = time.Duration(timeout) * time.Second
		}

		// Message
		message, _ := cmd.Flags().GetString("message")
		if len(message) > 0 {
			wh.SetRetryMessage(message)
		}

		// Quiet
		quiet, _ := cmd.Flags().GetBool("quiet")
		if quiet {
			wh.SetRetryMessage("")
		}

		//
		if len(wh.Scheme) == 0 || len(wh.Host) == 0 || wh.Port == 0 {
			os.Exit(1)
		} else {
			if err := wh.Wait(); err != nil {
				if err, ok := err.(*waithost.WaitHostError); ok {
					if err.Code() == waithost.TIMEOUT {
						os.Exit(1)
					}
				}
				os.Exit(2)
			}
		}

		// end
		os.Exit(0)
	},
}

func main() {

	cmd.Flags().StringP("host", "H", "", "Host or IP under test.")
	cmd.Flags().UintP("port", "p", 0, "TCP port under test.")
	cmd.Flags().UintP("delay", "d", 0, "Delay in seconds, before trying to contact the host.")
	cmd.Flags().StringP("message", "m", "Waiting for connection on {host}:{port}", "Retry message.")
	cmd.Flags().BoolP("quiet", "q", false, "Don't output any status messages.")
	cmd.Flags().UintP("timeout", "t", 0, "Timeout in seconds, zero for no timeout.")

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
