//go:generate go run -tags generate gen.go

package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	"log"

	"github.com/faroyam/gcr/gcrclient"
	"github.com/zserge/lorca"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	evalCh      = make(chan string)
	logChan     = make(chan string)
	msgChan     = make(chan string)
	conChan     = make(chan struct{})
	guiStopChan = make(chan struct{})

	opts    []grpc.DialOption
	address string
	port    string
	tls     bool

	c = gcrclient.NewClient()
)

func init() {
	flag.StringVar(&address, "a", "localhost", "server ip address")
	flag.StringVar(&port, "p", "50051", "port number")
	flag.BoolVar(&tls, "t", false, "enable tls ecryption")
	flag.Parse()
}

func main() {

	ui, err := lorca.New("", "", 360, 440)
	if err != nil {
		log.Fatalln(err)
	}
	defer ui.Close()

	ui.Bind("send", send)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()
	go http.Serve(ln, http.FileServer(FS))

	err = ui.Load(fmt.Sprintf("http://%s", ln.Addr()))
	if err != nil {
		log.Fatalln(err)
	}

	go guiUpdateListener(ui)
	go connect(address, port)
	<-guiStopChan
}

func connect(address, port string) {

	logChan <- fmt.Sprintf(`connecting to %v %v...`, address, port)
	msgChan <- fmt.Sprintf(`connecting to %v %v...`, address, port)

	onError := func(msg string, err error) {
		msgChan <- msg
		logChan <- fmt.Sprintf(`%s`, err.Error())
		<-time.After(3 * time.Second)
		conChan <- struct{}{}
	}

	if tls {
		creds, err := credentials.NewClientTLSFromFile(".cert.pem", "")
		if err != nil {
			onError(`failed to construct credentials from '.cert.pem'`, err)
			return
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	if err := c.SetClient(address, port, opts); err != nil {
		onError(`can't connect to the server`, err)
		return
	}
	logChan <- fmt.Sprintf(`connected`)

	if err := c.ReceiveName(); err != nil {
		onError(`can't connect to the server`, err)
		return
	}

	name := c.GetName()
	msgChan <- fmt.Sprintf(`connected as %s`, name)
	evalCh <- fmt.Sprintf(`msg.placeholder = '%s: '`, name)

	go msgReceiver()
	go infoReceiver()
}

func send(text string) {
	err := c.SendMessage(c.GetName(), text)
	if err != nil {
		logChan <- fmt.Sprintf(`%s`, err.Error())
		msgChan <- fmt.Sprintf(`can\'t send message`)
		return
	}
	logChan <- fmt.Sprintf(`send: %s`, text)
	evalCh <- fmt.Sprintf(`msg.value = ''`)
}

func msgReceiver() {
	for {
		author, text, err := c.ReceiveMessage()
		if err != nil {
			logChan <- fmt.Sprintf(`%s`, err.Error())
			msgChan <- fmt.Sprintf(`disconnected`)
			conChan <- struct{}{}
			return
		}
		logChan <- fmt.Sprintf(`received: %s: %s`, author, text)
		msgChan <- fmt.Sprintf(`%s: %s`, author, text)
	}
}

func infoReceiver() {
	for {
		count, err := c.ReceiveInfo()
		if err != nil {
			logChan <- fmt.Sprintf(`%s`, err.Error())
			return
		}
		logChan <- fmt.Sprintf(`received info: %d`, count)
		evalCh <- fmt.Sprintf(`counter.innerText = 'online users: %d'`, count)
	}
}

func guiUpdateListener(ui lorca.UI) {
	for {
		select {
		case txt := <-logChan:
			ui.Eval(fmt.Sprintf(`console.log('%s');`, gcrclient.Esc(txt)))
		case txt := <-msgChan:
			ui.Eval(fmt.Sprintf(`chat.append('%s\n'); chat.scrollTop = chat.scrollHeight;`, gcrclient.Esc(txt)))
		case txt := <-evalCh:
			ui.Eval(txt)
		case <-conChan:
			go connect(address, port)
		case <-ui.Done():
			guiStopChan <- struct{}{}
			return
		}
	}
}
