package publisher

import (
	"encoding/json"
	"goVncNet/helpers"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
	"github.com/tkanos/gonfig"
)

var redisDB *redis.Client

type RedisConfiguration struct {
	RedisHost  string
	RedisPort  string
	RedisDbNum int
}

// Chat server.
type Publisher struct {
	subscribers map[string]*Subscriber
	requests    map[string]string
	addCh       chan *Subscriber
	closeCh     chan string
	readCh      chan *Request
	doneCh      chan bool
}

func connectToRedis() (err error) {
	//

	redisConf := RedisConfiguration{}
	err = gonfig.GetConf("config/redis.json", &redisConf)
	if err != nil {
		//

		log.Fatalln("NO CONFIG FILE. ", err)
	}

	// var redisDBNumInt int = 1
	// redisHost, ok := os.LookupEnv("REDIS_PORT_6379_TCP_ADDR")
	// if !ok {
	// 	//
	//
	// 	redisHost = "0.0.0.0"
	// }
	//
	// redisPort, ok := os.LookupEnv("REDIS_PORT_6379_TCP_PORT")
	// if !ok {
	// 	//
	//
	// 	redisPort = "6379"
	// }
	//
	// redisDBNum, ok := os.LookupEnv("REDIS_PORT_6379_DB_NUM")
	// if !ok {
	// 	//
	//
	// 	redisDBNumInt = 1
	//
	// } else {
	// 	//
	//
	// 	if redisDBNumInt64, err := strconv.ParseInt(redisDBNum, 10, 64); err != nil {
	// 		//
	//
	// 		redisDBNumInt = int(redisDBNumInt64)
	//
	// 	} else {
	// 		//
	//
	// 		redisDBNumInt = 1
	// 	}
	// }

	redisDB = redis.NewClient(&redis.Options{
		Addr:         net.JoinHostPort(redisConf.RedisHost, redisConf.RedisPort),
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		DB:           redisConf.RedisDbNum,
	})

	statusCmd := redisDB.Ping()
	if helpers.IsRedisError(statusCmd) {
		//

		log.Fatalln("No connection to REDIS. ", statusCmd.Err())
		return statusCmd.Err()
	}

	return
}

// Create new chat server.
func NewPublisher() *Publisher {
	//

	connectToRedis()

	subscribers := make(map[string]*Subscriber)
	requests := make(map[string]string)
	addCh := make(chan *Subscriber)
	closeCh := make(chan string)
	readCh := make(chan *Request, 1024)
	doneCh := make(chan bool)

	return &Publisher{
		subscribers,
		requests,
		addCh,
		closeCh,
		readCh,
		doneCh,
	}
}

func (p *Publisher) Add(s *Subscriber) {
	//

	p.addCh <- s
}

func (p *Publisher) Done() {
	//

	p.doneCh <- true
}

// Listen and serve.
// It serves client connection and broadcast request.
var upgrader = websocket.Upgrader{} // use default options

func (p *Publisher) Status(w http.ResponseWriter, r *http.Request) {
	//

	log.Println("GET STATUS")
	w.WriteHeader(http.StatusOK)
}

func (p *Publisher) Serve(w http.ResponseWriter, r *http.Request) {
	//

	log.Println("NEW SERVE")

	log.Println("!")
	c, err := upgrader.Upgrade(w, r, nil)
	log.Println("!!")
	if err != nil {
		//

		log.Print("upgrade:", err)
		return
	}

	s := NewSubscriber(c, p.readCh, p.closeCh)
	p.Add(s)
	s.Listen()
}

func (p *Publisher) Listen() {
	//

	for {
		//

		select {

		case <-time.After(5 * time.Second):
			log.Println("TIMEOUT")
			for id, s := range p.subscribers {
				//

				log.Println("YOUR RESPONSE: " + id)
				redisCmd := redisDB.Get(p.requests[id])
				var raw map[string]string
				json.Unmarshal([]byte(redisCmd.Val()), &raw)
				s.Write(&Response{p.requests[id], raw})
			}

		// Add new a subscriber
		case s := <-p.addCh:
			log.Println("Added new subscriber")
			p.subscribers[s.id] = s
			log.Println("Now", len(p.subscribers), "subscribers connected.")

			// del a subscriber
		case id := <-p.closeCh:
			log.Println("Delete subscriber")

			p.subscribers[id].Del()
			delete(p.subscribers, id)

			// del a subscriber
		case r := <-p.readCh:
			log.Println("New request received: ", r)
			p.requests[r.id] = r.TYPE + ":" + r.SOURCE + ":" + r.VALUE

		case <-p.doneCh:
			log.Println("DONE")
			return
		}
	}
}
