package termhere

import (
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/creack/pty"
	"github.com/guoyk93/goyk/chdone"
	"github.com/guoyk93/rg"
	"github.com/guoyk93/termhere/thwire"
	"golang.org/x/term"
)

type ServerOptions struct {
	Token  string
	Listen string
}

func RunServer(opts ServerOptions) (err error) {
	defer rg.Guard(&err)

	log.Println("listening on:", opts.Listen)

	lis := rg.Must(net.Listen("tcp", opts.Listen))
	defer lis.Close()

	for {
		conn := rg.Must(lis.Accept())
		go serverHandleConnection(conn, opts.Token)
	}
}

var (
	occupied     bool
	occupiedLock sync.Locker = &sync.Mutex{}
)

func serverOccupy() bool {
	occupiedLock.Lock()
	defer occupiedLock.Unlock()
	if occupied {
		return false
	}
	occupied = true
	return true
}

func serverUnoccupy() {
	occupiedLock.Lock()
	defer occupiedLock.Unlock()
	occupied = false
}

func serverExposeEnv() map[string]string {
	env := map[string]string{}
	for _, key := range []string{
		"TERM",
	} {
		if val := os.Getenv(key); val != "" {
			env[key] = val
		}
	}
	return env
}

func serverHandleConnection(conn net.Conn, token string) {
	defer conn.Close()

	var err error

	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	defer func() {
		f := thwire.Frame{Kind: thwire.KindExit}
		if err != nil {
			f.Exit.Code = 1
			f.Exit.Message = []byte(err.Error())
			log.Println("error:", err)
		}
		_ = gw.Encode(f)
	}()
	rg.Guard(&err)

	log.Println("client authenticating")

	// read auth frame
	af := thwire.Frame{}
	rg.Must0(gr.Decode(&af))
	rg.Must0(thwire.ValidateAuthFrame(af, token))

	// send auth frame
	af = thwire.Frame{}
	af.Auth.Env = serverExposeEnv()
	rg.Must0(thwire.CreateAuthFrame(&af, token))
	rg.Must0(gw.Encode(af))

	log.Println("client authenticated")

	if !serverOccupy() {
		log.Println("occupied")
		err = errors.New("occupied")
		return
	}
	defer serverUnoccupy()

	var (
		chSig      = make(chan os.Signal, 1)
		chOutgoing = make(chan thwire.Frame)
		chIncoming = make(chan thwire.Frame)

		done = chdone.New()
	)

	signal.Notify(chSig, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(chSig)

	go func() {
		for {
			select {
			case _sig := <-chSig:
				var f *thwire.Frame

				if sig, ok := _sig.(syscall.Signal); ok {
					if sig == syscall.SIGWINCH {
						if size, err := pty.GetsizeFull(os.Stdin); err == nil {
							f = &thwire.Frame{
								Kind: thwire.KindResize,
								Resize: thwire.FrameResize{
									Rows: size.Rows,
									Cols: size.Cols,
									X:    size.X,
									Y:    size.Y,
								},
							}
						}
					} else {
						f = &thwire.Frame{
							Kind: thwire.KindSignal,
							Signal: thwire.FrameSignal{
								Number: int(sig),
							},
						}
					}
				}

				if f != nil {
					select {
					case chOutgoing <- *f:
					case <-done.C:
						return
					}
				}
			case <-done.C:
				return
			}
		}
	}()

	// drain incoming frames
	go func() {
		for {
			var f thwire.Frame
			if err := gr.Decode(&f); err != nil {
				if done.TryClose() {
					if err == io.EOF {
						err = nil
						log.Println("client exited")
					} else {
						log.Println("read frame error:", err)
					}
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

	// drain stdin
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				if done.TryClose() {
					log.Println("read stdin error:", err)
				}
				return
			}
			f := thwire.Frame{
				Kind: thwire.KindStdin,
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

	oldState := rg.Must(term.MakeRaw(int(os.Stdin.Fd())))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	oldStateOut := rg.Must(term.MakeRaw(int(os.Stdout.Fd())))
	defer term.Restore(int(os.Stdout.Fd()), oldStateOut)

	chSig <- syscall.SIGWINCH

	for {
		select {
		case f := <-chIncoming:
			switch f.Kind {
			case thwire.KindIdle:
				// ignore idle
			case thwire.KindSignal, thwire.KindStdin, thwire.KindResize:
				if done.TryClose() {
					log.Println("invalid client frame:", f.Kind.String())
				}
			case thwire.KindStdout, thwire.KindStderr:
				_, _ = os.Stdout.Write(f.Data)
			case thwire.KindExit:
				log.Println("client exited:", f.Exit.Code, string(f.Exit.Message))
				done.TryClose()
			}
		case <-done.C:
			return
		}
	}

}
