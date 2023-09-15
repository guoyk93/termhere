package termhere

import (
	"context"
	"encoding/gob"
	"fmt"
	"github.com/guoyk93/uniconn"
	"io"
	"log"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/guoyk93/goyk/chdone"
	"github.com/guoyk93/rg"
	"github.com/guoyk93/termhere/thwire"
)

type ClientOptions struct {
	Token    string
	Server   string
	Command  []string
	CAFile   string
	Insecure bool
}

func RunClient(opts ClientOptions) (err error) {
	defer rg.Guard(&err)

	log.Println("connecting to:", opts.Server)

	var cfg uniconn.DialConfig
	if cfg, err = uniconn.ParseDialURI(opts.Server, map[string]string{
		uniconn.OptionCAFile:   opts.CAFile,
		uniconn.OptionInsecure: fmt.Sprintf("%v", opts.Insecure),
	}); err != nil {
		return
	}

	conn := rg.Must(cfg.Dial(context.Background()))
	defer conn.Close()

	log.Println("authenticating")

	// gob reader/writer
	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	// send auth frame
	af := thwire.Frame{}
	rg.Must0(thwire.CreateAuthFrame(&af, opts.Token))
	rg.Must0(gw.Encode(af))

	// read auth frame
	af = thwire.Frame{}
	rg.Must0(gr.Decode(&af))
	rg.Must0(thwire.ValidateAuthFrame(af, opts.Token))

	log.Println("authenticated")

	log.Println("executing:", strings.Join(opts.Command, " "))

	// command
	cmd := exec.Command(opts.Command[0], opts.Command[1:]...)
	cmd.Env = os.Environ()
	for k, v := range af.Auth.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	defer func() {
		if s := cmd.ProcessState; s != nil {
			log.Println("command exited:", s.String())
			_ = gw.Encode(thwire.Frame{
				Kind: thwire.KindExit,
				Exit: thwire.FrameExit{
					Code:    s.ExitCode(),
					Message: []byte(s.String()),
				},
			})
		}
	}()

	// start pty
	pt := rg.Must(pty.Start(cmd))
	defer pt.Close()

	var (
		chIncoming = make(chan thwire.Frame)
		chOutgoing = make(chan thwire.Frame)

		done = chdone.New()
	)

	// drain incoming frames
	go func() {
		for {
			var f thwire.Frame
			if err := gr.Decode(&f); err != nil {
				if done.TryClose() {
					log.Println("read frame error:", err)
				}
				return
			}
			select {
			case chIncoming <- f:
			case <-done.C:
				return
			}
		}
	}()

	// drain outgoing frames
	go func() {
		for {
			select {
			case f := <-chOutgoing:
				if err := gw.Encode(f); err != nil {
					if done.TryClose() {
						log.Println("write frame error:", err)
					}
					return
				}
			case <-done.C:
				return
			}
		}
	}()

	// drain process pty out
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := pt.Read(buf)
			if err != nil {
				if done.TryClose() {
					if err == io.EOF {
						log.Println("command exited")
					} else {
						log.Println("read pty error:", err)
					}
				}
				return
			}
			f := thwire.Frame{
				Kind: thwire.KindStdout,
				Data: slices.Clone(buf[:n]),
			}
			select {
			case chOutgoing <- f:
			case <-done.C:
				return
			}
		}
	}()

	// periodic idle
	go func() {
		tk := time.NewTicker(10 * time.Second)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				f := thwire.Frame{Kind: thwire.KindIdle}
				select {
				case chOutgoing <- f:
				case <-done.C:
					return
				}
			case <-done.C:
				return
			}
		}
	}()

	for {
		select {
		case f := <-chIncoming:
			switch f.Kind {
			case thwire.KindIdle:
				// ignore idle
			case thwire.KindSignal:
				if process := cmd.Process; process != nil {
					_ = process.Signal(syscall.Signal(f.Signal.Number))
				}
			case thwire.KindStdin:
				if _, err = pt.Write(f.Data); err != nil {
					if done.TryClose() {
						log.Println("write stdin error:", err)
					}
				}
			case thwire.KindStdout, thwire.KindStderr:
				if done.TryClose() {
					log.Println("invalid server frame:", f.Kind.String())
				}
			case thwire.KindExit:
				if done.TryClose() {
					if f.Exit.Code == 0 {
						log.Println("server exited:", string(f.Exit.Message))
					} else {
						log.Println("server error:", string(f.Exit.Message))
					}
				}
			case thwire.KindResize:
				_ = pty.Setsize(pt, &pty.Winsize{
					Rows: f.Resize.Rows,
					Cols: f.Resize.Cols,
					X:    f.Resize.X,
					Y:    f.Resize.Y,
				})
			}
		case <-done.C:
			return
		}
	}
}
