package publisher

import (
	"encoding/hex"
	"log"

	"github.com/gorilla/websocket"
	"golang.org/x/crypto/blake2b"
)

const channelBufSize = 100

// Chat Subscriber.
type Subscriber struct {
	id      string
	conn    *websocket.Conn
	writeCh chan *Response
	readCh  chan *Request
	closeCh chan string
	doneCh  chan bool
}

// Create new chat Subscriber.
func NewSubscriber(conn *websocket.Conn, readCh chan *Request, closeCh chan string) *Subscriber {
	//

	if conn == nil {
		//

		panic("conn cannot be nil")
	}

	hash, _ := blake2b.New256([]byte(conn.RemoteAddr().String()))
	id := hex.EncodeToString(hash.Sum(nil))
	writeCh := make(chan *Response, channelBufSize)
	doneCh := make(chan bool)

	return &Subscriber{id, conn, writeCh, readCh, closeCh, doneCh}
}

func (s *Subscriber) Conn() *websocket.Conn {
	//

	return s.conn
}

func (s *Subscriber) Write(r *Response) {
	//

	s.writeCh <- r
}

func (s *Subscriber) Del() {
	//

	log.Println("Deleting subscriber")
	s.doneCh <- true
	s.conn.Close()
}

// Listen Write and Read request via chanel
func (s *Subscriber) Listen() {
	//

	log.Println("Liesten to subscriber")
	go s.listenWrite()
	s.listenRead()
}

// Listen write request via chanel
func (s *Subscriber) listenWrite() {
	//

	log.Println("Listening write to Subscriber")
	for {
		//

		select {

		// send message to the Subscriber
		case response := <-s.writeCh:
			log.Println("Send:", response)
			s.conn.WriteJSON(response)

			// receive done request
		case <-s.doneCh:
			return
		}
	}
}

// Listen read request via chanel
func (s *Subscriber) listenRead() {
	//

	log.Println("Listening read from Subscriber")
	for {
		//

		select {

		// receive done request
		case <-s.doneCh:
			return

		// read data from websocket connection
		default:
			var orf OracleRequestFormat
			err := s.conn.ReadJSON(&orf)
			if err != nil {
				//

				log.Println(err)
				s.closeCh <- s.id
				return
			}

			s.readCh <- &Request{s.id, orf}
		}
	}
}
