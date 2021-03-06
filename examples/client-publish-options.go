// +build ignore

package main

import (
	"fmt"
	"net"
	"time"

	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/pkg/rtph264"
)

// This example shows how to
// 1. set additional client options
// 2. generate RTP/H264 frames from a file with Gstreamer
// 3. connect to a RTSP server, announce a H264 track
// 4. write the frames to the server

func main() {
	// open a listener to receive RTP/H264 frames
	pc, err := net.ListenPacket("udp4", "127.0.0.1:9000")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	fmt.Println("Waiting for a rtp/h264 stream on port 9000 - you can send one with gstreamer:\n" +
		"gst-launch-1.0 filesrc location=video.mp4 ! qtdemux ! video/x-h264" +
		" ! h264parse config-interval=1 ! rtph264pay ! udpsink host=127.0.0.1 port=9000")

	// wait for RTP/H264 frames
	decoder := rtph264.NewDecoderFromPacketConn(pc)
	sps, pps, err := decoder.ReadSPSPPS()
	if err != nil {
		panic(err)
	}
	fmt.Println("stream connected")

	// create a H264 track
	track, err := gortsplib.NewTrackH264(96, sps, pps)
	if err != nil {
		panic(err)
	}

	// ClientConf allows to set additional client options
	conf := gortsplib.ClientConf{
		// the stream protocol (UDP or TCP). If nil, it is chosen automatically
		StreamProtocol: nil,
		// timeout of read operations
		ReadTimeout: 10 * time.Second,
		// timeout of write operations
		WriteTimeout: 10 * time.Second,
	}

	// connect to the server and start publishing the track
	conn, err := conf.DialPublish("rtsp://localhost:8554/mystream",
		gortsplib.Tracks{track})
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, 2048)
	for {
		// read frames from the source
		n, _, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		// write track frames
		err = conn.WriteFrame(track.ID, gortsplib.StreamTypeRtp, buf[:n])
		if err != nil {
			panic(err)
		}
	}
}
