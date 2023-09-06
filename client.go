package termhere

import (
	"encoding/gob"
	"github.com/creack/pty"
	"github.com/guoyk93/rg"
	"github.com/guoyk93/termhere/thdone"
	"github.com/guoyk93/termhere/thwire"
	"log"
	"net"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"time"
)

type ClientOptions struct {
	Token   string
	Server  string
	Command []string
}

func RunClient(opts ClientOptions) (err error) {
	defer rg.Guard(&err)

	log.Println("connecting to:", opts.Server)

	conn := rg.Must(net.Dial("tcp", opts.Server))
	defer conn.Close()

	log.Println("authenticating")

	// gob reader/writer
	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	// send auth frame
	var af thwire.Frame
	rg.Must0(thwire.CreateAuthFrame(&af, opts.Token))
	rg.Must0(gw.Encode(af))

	// read auth frame
	rg.Must0(gr.Decode(&af))
	rg.Must0(thwire.ValidateAuthFrame(af, opts.Token))

	log.Println("authenticated")

	log.Println("executing:", strings.Join(opts.Command, " "))

	// command
	cmd := exec.Command(opts.Command[0], opts.Command[1:]...)

	// start pty
	pt := rg.Must(pty.Start(cmd))
	defer pt.Close()

	var (
		chIncoming = make(chan thwire.Frame)
		chOutgoing = make(chan thwire.Frame)

		done = thdone.New()
	)

	// drain incoming frames
	go func() {
		for {
			var f thwire.Frame
			if err := gr.Decode(&f); err != nil {
				log.Println("read frame error:", err)
				done.Close()
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
					log.Println("write frame error:", err)
					done.Close()
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
				log.Println("read pty error:", err)
				done.Close()
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
					log.Println("write stdin error:", err)
					done.Close()
				}
			case thwire.KindStdout, thwire.KindStderr:
				log.Println("invalid server frame:", f.Kind.String())
				done.Close()
			case thwire.KindError:
				log.Println("server error:", string(f.Data))
				done.Close()
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
