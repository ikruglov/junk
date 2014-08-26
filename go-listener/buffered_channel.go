package main

// Implementation of unlimited buffered channels
// http://rogpeppe.wordpress.com/2010/02/10/unlimited-buffering-with-low-overhead/

func MakeBufferedChannel() (chan<- []byte, <-chan []byte) {
	in := make(chan []byte, 100)
	out := make(chan []byte)

	go func() {
		var i = 0 // location of first value in buffer.
		var n = 0 // number of items in buffer.
		var outc chan<- []byte
		var buf = make([][]byte, 8)

		for {
			select {
			case e, ok := <-in:
				if ok {
					j := i + n
					if j >= len(buf) {
						j -= len(buf)
					}

					buf[j] = e

					if n++; n == len(buf) {
						// buffer full: expand it
						b := make([][]byte, n*2)
						copy(b, buf[i:])
						copy(b[n-i:], buf[0:i])
						i = 0
						buf = b
					}

					outc = out
				} else {
					in = nil
					if n == 0 {
						close(out)
						return
					}
				}

			case outc <- buf[i]:
				buf[i] = nil
				if i++; i == len(buf) {
					i = 0
				}

				if n--; n == 0 {
					// buffer empty: don't try to send on output
					if in == nil {
						close(out)
						return
					}

					outc = nil
				}

				if len(buf) > 128 && n*3 < len(buf) {
					// buffer too big, shrink it
					b := make([][]byte, len(buf)/2)
					j := i + n
					if j > len(buf) {
						// wrap around
						k := j - len(buf)
						j = len(buf)
						copy(b, buf[i:j])
						copy(b[j-i:], buf[0:k])
					} else {
						// contiguous
						copy(b, buf[i:j])
					}

					i = 0
					buf = b
				}
			}
		}
	}()

	return in, out
}
