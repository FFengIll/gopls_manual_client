package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/mox692/gopls_manual_client/protocol"
	"github.com/sourcegraph/go-langserver/langserver/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

/**
 * these struct is referenced from  golang/tools/internal/lsp/protocol/tsprotocol.go
 * ref: https://github.com/golang/tools/blob/master/internal/lsp/protocol/tsprotocol.go
 */
type clientHandler struct {
}

func (c *clientHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	fmt.Printf("called!! request is : %+v\n", req)
	if !req.Notif {
		return
	}

	switch req.Method {
	case "textDocument/documentHighlight":
		var params lsp.PublishDiagnosticsParams
		b, _ := req.Params.MarshalJSON()
		json.Unmarshal(b, &params)
	}
}

func main() {
	var (
		port = flag.String("port", "37374", "gopls's port")
	)
	flag.Parse()

	c, err := net.Dial("tcp", ":"+*port)
	if err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	conn := jsonrpc2.NewConn(context.Background(), jsonrpc2.NewBufferedStream(c, jsonrpc2.VSCodeObjectCodec{}), &clientHandler{})

	fullpath, err := filepath.Abs(".")
	uri := util.PathToURI(fullpath)
	if err != nil {
		panic(err)
	}

	params := protocol.InitializeParams{
		RootPath: fullpath,
		RootURI:  protocol.DocumentURI(string(uri)),
	}

	var initializeRes interface{}
	err = conn.Call(context.Background(), "initialize", params, &initializeRes)

	b, err := json.Marshal(initializeRes)

	if err != nil {
		panic(err)
	}
	fmt.Println()
	fmt.Println(string(b))

	fmt.Println("initial Call done!! next handshake...")

	// next, send handShake method...
	time.Sleep(time.Second * 1)
	handshakeReq := protocol.HandshakeRequest{}
	handshakeRes := protocol.HandshakeResponse{}
	err = conn.Call(context.Background(), "gopls/handshake", handshakeReq, &handshakeRes)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("handshake result: %+v\n", handshakeRes)

	// wait for message from server...
	ctx, cancel := context.WithCancel(context.Background())
	go waitReq(ctx)

	time.Sleep(time.Second * 1)

	// req didchange to server...
	var didChangeResult interface{}
	didChangeParams := protocol.DidChangeTextDocumentParams{
		TextDocument: protocol.VersionedTextDocumentIdentifier{
			Version: 2,
			TextDocumentIdentifier: protocol.TextDocumentIdentifier{
				URI: "dammyURI",
			},
		},
	}
	err = conn.Call(context.Background(), "textDocument/didChange", didChangeParams, &didChangeResult)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("didChange Result: %+v\n", didChangeResult)
	cancel()

	fmt.Println("cancel done!! program exit...")
}

func waitReq(ctx context.Context) {
	for {
		time.Sleep(time.Second * 1)
		select {
		case <-ctx.Done():
			fmt.Printf("waitReq Done...\n")
			return
		default:
		}
	}
}
