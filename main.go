package main

import (
	"fmt"
	"github.com/fanout/go-gripcontrol"
	pubcontrol "github.com/fanout/go-pubcontrol"
	"github.com/gin-gonic/gin"
	zmq "github.com/pebbe/zmq4"
	"io"
	"io/ioutil"
	"net/http"
)

func main() {
	router := gin.Default()

	zctx, err := zmq.NewContext()
	s, err := zctx.NewSocket(zmq.PUB)
	err = s.Bind("tcp://*:8081")
	if err != nil {
		fmt.Println("Err binding socket : ", err.Error())
	}

	router.Handle("GET", "/publish-on-zmq", func(context *gin.Context) {
		fmt.Println("Got /publish-on-zmq request on origin server")
		PublishOverZMQ(context.Writer, context.Request, s)
	})
	router.Handle("GET", "/stats", func(context *gin.Context) {
		fmt.Println("Got /stats request on origin server")
		GetStats(context.Writer, context.Request)
	})
	router.Handle("POST", "/subscribe", func(context *gin.Context) {
		fmt.Println("Got /subscribe  request on origin server")
		Subscribe(context.Writer, context.Request)
	})
	router.Handle("POST", "/publish", func(context *gin.Context) {
		fmt.Println("About to publish")
		Publish(context.Writer, context.Request)
	})

	err1 := http.ListenAndServe(":8080", router)
	if err1 != nil {
		fmt.Println("Error starting router : ", err1.Error())
	}
}

func PublishOverZMQ(writer http.ResponseWriter, request *http.Request, s *zmq.Socket) {
	fmt.Println("/publish-on-zmq got ==> Body : ")

	sentTopic, err := s.Send(`publish`, zmq.SNDMORE)
	if err != nil {
		fmt.Println("Err sending : ", err.Error())
	}
	fmt.Printf("sentTopic : ", sentTopic)

	sent, err := s.Send(`J{"id": "1", "channel": "test", "formats": {"ws-message": {"content": "blah\n"}}}`, 0)
	if err != nil {
		fmt.Println("Err sending : ", err.Error())
	}
	fmt.Printf("Sent : ", sent)
}

func GetStats(writer http.ResponseWriter, request *http.Request) {
	zctx, err := zmq.NewContext()
	if err != nil {
		fmt.Println("Err ctx : ", err.Error())
	}
	socket, err := zctx.NewSocket(zmq.REQ)
	if err != nil {
		fmt.Println("Err soc : ", err.Error())
	}
	err = socket.Connect("")

	for {
		recved, err := socket.Recv(0)
		if err != nil {
			fmt.Println("Err receiving : ", err.Error())
		}
		fmt.Println("Received : ", recved)
	}
}

func Subscribe(writer http.ResponseWriter, request *http.Request) {

	inputBody, _ := ioutil.ReadAll(request.Body)
	fmt.Println("/publish-on-zmq got ==> Body : ", string(inputBody))

	writer.Header().Set("Content-Type", "application/websocket-events")

	inEvents, err := gripcontrol.DecodeWebSocketEvents(string(inputBody))
	if err != nil {
		panic("Failed to decode WebSocket inEvents: " + err.Error())
	}

	fmt.Printf("IN events.length: %d", len(inEvents))
	fmt.Printf("IN events[0]: %+v", inEvents[0])

	shouldSubscribe := false

	if inEvents[0].Type == "OPEN" {
		fmt.Println("OPEN event came to origin server ! ")
		writer.Header().Set("Sec-WebSocket-Extensions", `grip; message-prefix=""`)
		shouldSubscribe = true
	}

	fmt.Println("should subscribe ? ", shouldSubscribe)

	if shouldSubscribe {
		wsControlMessage, err := gripcontrol.WebSocketControlMessage("subscribe",
			map[string]interface{}{"channel": "test"})
		if err != nil {
			panic("Unable to create control message: " + err.Error())
		}

		outEvents := []*gripcontrol.WebSocketEvent{
			&gripcontrol.WebSocketEvent{Type: "OPEN"},
			&gripcontrol.WebSocketEvent{Type: "TEXT",
				Content: "c:" + wsControlMessage},
		}

		n, err1 := io.WriteString(writer, gripcontrol.EncodeWebSocketEvents(outEvents))
		if err1 != nil {
			fmt.Println("Err writing outEvents to writer: ", err1.Error())
			return
		}
		fmt.Println("Wrote : ", n)
	}

}

func Publish(writer http.ResponseWriter, request *http.Request) {

	data, _ := ioutil.ReadAll(request.Body)
	fmt.Println("Data to be published: ", string(data))

	writer.Header().Set("Content-Type", "application/websocket-events")

	pub := gripcontrol.NewGripPubControl([]map[string]interface{}{
		map[string]interface{}{"control_uri": "http://localhost:5561"}})

	format := &gripcontrol.WebSocketMessageFormat{
		Content: data}

	item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")

	err := pub.Publish("test", item)

	if err != nil {
		panic("Publish failed with: " + err.Error())
	}

}
