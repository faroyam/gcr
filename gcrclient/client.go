package gcrclient

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/faroyam/gcr/gcr"
	"google.golang.org/grpc"
)

// Client ...
type Client struct {
	sync.RWMutex
	dial       *grpc.ClientConn
	connection gcr.ChatRoomClient
	msgStream  gcr.ChatRoom_BroadcastClient
	infoStream gcr.ChatRoom_InfoClient
	name       string
}

// SetClient connects to rpc server
func (c *Client) SetClient(address string, port string, opts []grpc.DialOption) (err error) {
	dial, err := grpc.Dial(address+":"+port, opts...)
	if err != nil {
		return
	}

	c.dial = dial
	c.connection = gcr.NewChatRoomClient(dial)

	ctx := context.Background()

	msgStream, err := c.connection.Broadcast(ctx)
	if err != nil {
		return
	}
	c.msgStream = msgStream

	infoStream, err := c.connection.Info(ctx, &gcr.InfoRequest{})
	if err != nil {
		return
	}
	c.infoStream = infoStream

	return
}

// SendMessage sends  messages to rpc stream
func (c *Client) SendMessage(author, message string) error {
	if err := c.msgStream.Send(&gcr.Message{Author: author, Text: message}); err != nil {
		return err
	}
	return nil
}

// ReceiveMessage reads messages from rpc stream
func (c *Client) ReceiveMessage() (string, string, error) {
	in, err := c.msgStream.Recv()
	if err != nil {
		return "", "", err
	}
	return in.Author, in.Text, nil
}

// ReceiveName sends rpc NameRequets message
func (c *Client) ReceiveName() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	r, err := c.connection.RandName(ctx, &gcr.NameRequets{})
	if err != nil {
		return err
	}
	c.name = r.Name
	return nil
}

// ReceiveInfo reads information from rpc stream
func (c *Client) ReceiveInfo() (int64, error) {
	info, err := c.infoStream.Recv()
	if err != nil {
		return 0, err
	}
	return info.ClientsCount, err
}

// GetName returns client name
func (c *Client) GetName() string {
	c.RLock()
	defer c.RUnlock()
	return c.name
}

// NewClient returns new grpc client instance
func NewClient() *Client {
	return &Client{}
}

// Esc replaces ' by \' in user input
func Esc(s string) string {
	return strings.Replace(s, `'`, `\'`, -1)
}
