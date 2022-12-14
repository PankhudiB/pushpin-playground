package main

import (
	"fmt"
	"github.com/fanout/go-gripcontrol"
	"github.com/fanout/go-pubcontrol"
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"time"

	"net/http"
)

func main() {
	router := gin.Default()

	router.Handle("GET", "/data", func(context *gin.Context) {
		fmt.Printf("REQ: %+v", context.Request)
		fmt.Println("REQ.Header : ", context.Request.Header)
		fmt.Println("Received request ...")

		context.Writer.Header().Set("Content-Type", "text/plain")
		context.Writer.Header().Set("Grip-Hold", "stream")
		context.Writer.Header().Set("Grip-Channel", "test")
		context.Status(http.StatusOK)
		return
	})

	router.Handle("GET", "/", func(context *gin.Context) {
		HandleRequest(context.Writer, context.Request)
	})

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Error starting router : ", err.Error())
	}
}

func HandleRequest(writer http.ResponseWriter, request *http.Request) {
	// Validate the Grip-Sig header:
	//if !gripcontrol.ValidateSig(request.Header["Grip-Sig"][0], "<key>") {
	//	http.Error(writer, "GRIP authorization failed", http.StatusUnauthorized)
	//	return
	//}

	// Set the headers required by the GRIP proxy:
	writer.Header().Set("Sec-WebSocket-Extensions", "grip; message-prefix=\"\"")
	writer.Header().Set("Content-Type", "application/websocket-events")
	// Decode the incoming WebSocket events:
	body, _ := ioutil.ReadAll(request.Body)
	fmt.Println("Body : ", string(body))

	inEvents, err := gripcontrol.DecodeWebSocketEvents(string(body))
	if err != nil {
		panic("Failed to decode WebSocket events: " + err.Error())
	}

	fmt.Printf("IN events.length: %d", len(inEvents))
	fmt.Printf("IN events[0]: %+v", inEvents[0])

	if inEvents[0].Type == "OPEN" {
		// Create the WebSocket control message:
		wsControlMessage, err := gripcontrol.WebSocketControlMessage("subscribe",
			map[string]interface{}{"channel": "ws-test"})
		if err != nil {
			panic("Unable to create control message: " + err.Error())
		}

		// Open the WebSocket and subscribe it to a channel:
		outEvents := []*gripcontrol.WebSocketEvent{
			&gripcontrol.WebSocketEvent{Type: "OPEN"},
			&gripcontrol.WebSocketEvent{Type: "TEXT",
				Content: "c:" + wsControlMessage}}
		_, err1 := io.WriteString(writer, gripcontrol.EncodeWebSocketEvents(outEvents))
		if err1 != nil {
			fmt.Println("Err writing: ", err1.Error())
		}

		fmt.Println("Abount to enter goroutine ...")
		go func() {
			// Wait 3 seconds and publish a message to the subscribed channel:
			fmt.Println("Waiting for 10 seconds.......")
			time.Sleep(10 * time.Second)

			pub := gripcontrol.NewGripPubControl([]map[string]interface{}{
				map[string]interface{}{"control_uri": "http://pushpin:5561"}})
			format := &gripcontrol.WebSocketMessageFormat{
				Content: []byte("Test WebSocket Publish!!")}
			item := pubcontrol.NewItem([]pubcontrol.Formatter{format}, "", "")
			err = pub.Publish("ws-test", item)
			if err != nil {
				panic("Publish failed with: " + err.Error())
			}
		}()
	}
}
