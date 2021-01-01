package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"testing"

	"github.com/kr/pty"
	"github.com/pion/webrtc/v3"
)

func TestHosttDataChannelOnMessage(t *testing.T) {
	hs := hostSession{ptmxReady: true}
	hs.errChan = make(chan error, 1)
	onMessage := hs.dataChannelOnMessage()
	quitPayload := webrtc.DataChannelMessage{IsString: true, Data: []byte("quit")}
	onMessage(quitPayload)

	select {
	case err := <-hs.errChan:
		if err != nil {
			t.Error(err)
		}
	default:
		t.Error(errors.New("should return errChan"))
	}

	stdoutMock := tmpFile()
	hs.ptmx = stdoutMock

	binaryPayload := webrtc.DataChannelMessage{IsString: false, Data: []byte("s")}
	onMessage(binaryPayload)
	stdoutMock.Seek(0, 0)
	msg, _ := ioutil.ReadAll(stdoutMock)
	if string(msg) != "s" {
		t.Error("bytes not written to stdout")
	}

}

func makeShPty(t *testing.T) (func(p webrtc.DataChannelMessage), hostSession) {
	hs := hostSession{ptmxReady: true}
	hs.errChan = make(chan error, 1)
	onMessage := hs.dataChannelOnMessage()
	c := exec.Command("sh")
	var err error
	// redefine the global ptmx
	hs.ptmx, err = pty.Start(c)
	if err != nil {
		t.Error(err)
	}
	return onMessage, hs
}

func TestClientSetSizeOnMessage(t *testing.T) {
	onMessage, hs := makeShPty(t)

	sizeOnlyPayload := webrtc.DataChannelMessage{IsString: true, Data: []byte(`["set_size", 20, 30]`)}
	onMessage(sizeOnlyPayload)

	size, err := pty.GetsizeFull(hs.ptmx)
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%v", size) != "&{20 30 0 0}" {
		t.Error("wrong size", size)
	}

	sizeAndCursorPayload := webrtc.DataChannelMessage{IsString: true, Data: []byte(`["set_size", 20, 30, 10, 11]`)}
	onMessage(sizeAndCursorPayload)

	size, err = pty.GetsizeFull(hs.ptmx)
	if err != nil {
		t.Error(err)
	}
	if fmt.Sprintf("%v", size) != "&{20 30 10 11}" {
		t.Error("wrong size", size)
	}
	select {
	case err := <-hs.errChan:
		if err != nil {
			t.Error(err)
		}
	default:

	}
}
