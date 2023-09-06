package termhere

import (
	"encoding/gob"
	"errors"
	"github.com/creack/pty"
	"github.com/guoyk93/rg"
	"github.com/guoyk93/termhere/thdone"
	"github.com/guoyk93/termhere/thwire"
	"golang.org/x/term"
	"log"
	"net"
	"os"
	"os/signal"
	"slices"
	"sync"
	"syscall"
	"time"
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

func serverHandleConnection(conn net.Conn, token string) {
	defer conn.Close()

	var err error

	gr := gob.NewDecoder(conn)
	gw := gob.NewEncoder(conn)

	defer func() {
		if err == nil {
			return
		}
		log.Println("error:", err)
		f := thwire.Frame{Kind: thwire.KindError, Data: []byte(err.Error())}
		_ = gw.Encode(f)
	}()
	rg.Guard(&err)

	log.Println("authenticating")

	// read auth frame
	var af thwire.Frame
	rg.Must0(gr.Decode(&af))
	rg.Must0(thwire.ValidateAuthFrame(af, token))

	// send auth frame
	rg.Must0(thwire.CreateAuthFrame(&af, token))
	rg.Must0(gw.Encode(af))

	log.Println("authenticated")

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

		done = thdone.New()
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

	// drain stdin
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				log.Println("read stdin error:", err)
				done.Close()
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

	for {
		select {
		case f := <-chIncoming:
			switch f.Kind {
			case thwire.KindIdle:
				// ignore idle
			case thwire.KindSignal, thwire.KindStdin, thwire.KindResize:
				log.Println("invalid client frame:", f.Kind.String())
				done.Close()
			case thwire.KindStdout, thwire.KindStderr:
				_, _ = os.Stdout.Write(f.Data)
			case thwire.KindError:
				log.Println("client error:", string(f.Data))
				done.Close()
			}
		case <-done.C:
			return
		}
	}

}
