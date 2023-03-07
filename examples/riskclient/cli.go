package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"

	"github.com/frozenpine/rhmonitor4go/service"
)

type command struct {
	client service.RohonMonitorClient
	cmd    reflect.Method
	// args   []interface{}
	args []string
}

func (cmd *command) Execute() error {
	log.Printf("Executing %s with args: %v", cmd.cmd.Name, cmd.args)

	return nil
}

var cmdCache = sync.Pool{New: func() any { return &command{} }}

const cmdPrefix = "> "

type CLI struct {
	ctx     context.Context
	cancel  context.CancelFunc
	start   sync.Once
	stop    sync.Once
	client  service.RohonMonitorClient
	methods map[string]reflect.Method
}

func (cli *CLI) quit() {
	cli.stop.Do(func() {
		cli.cancel()
	})
}

func (cli *CLI) newCMD(cmd string, args ...string) *command {
	method, exist := cli.methods[cmd]
	if !exist {
		return nil
	}

	c := cmdCache.Get().(*command)
	runtime.SetFinalizer(c, func(obj interface{}) {
		cmdCache.Put(obj)
	})

	c.client = cli.client
	c.cmd = method

	c.args = args

	return c
}

func (cli *CLI) cmdLoop() {
	cmdRd := bufio.NewReader(os.Stdin)
	cmdCh := make(chan string)

	go func() {
		defer close(cmdCh)

		os.Stdin.WriteString(cmdPrefix)

		for {
			input, err := cmdRd.ReadString('\n')

			if err != nil {
				if err == io.EOF {
					break
				}

				log.Printf("Read input failed: %+v", err)
				continue
			}

			if input == "quit" {
				break
			}

			cmdCh <- input

			os.Stdin.WriteString(cmdPrefix)
		}
	}()

	for {
		select {
		case <-cli.ctx.Done():
			os.Stdin.Close()
			cli.quit()
			return
		case input := <-cmdCh:
			if input == "" {
				continue
			}

			buffer := make([]string, 0, 2)

			for _, sec := range strings.Split(input, " ") {
				v := strings.Trim(sec, " \t")

				if v == "" {
					continue
				}

				buffer = append(buffer, v)
			}

			if len(buffer) < 1 {
				log.Printf("Invalid command.")
				continue
			}

			cmd := cli.newCMD(buffer[0], buffer[1:]...)
			if cmd == nil {
				log.Printf("Command[%s] not exist", buffer[0])
				continue
			}

			if err := cmd.Execute(); err != nil {
				log.Printf("Command execution failed: %+v", err)
			}
		}
	}
}

func (cli *CLI) Serve(ctx context.Context, client service.RohonMonitorClient) error {
	if client == nil {
		return errors.New("client conn missing")
	}

	cli.start.Do(func() {
		if ctx == nil {
			ctx = context.Background()
		}

		cli.ctx, cli.cancel = context.WithCancel(ctx)

		clientType := reflect.TypeOf(client)

		for i := 0; i < clientType.NumMethod(); i++ {
			method := clientType.Method(i)
			cli.methods[method.Name] = method
		}

		go cli.cmdLoop()
	})

	<-cli.ctx.Done()

	return nil
}
