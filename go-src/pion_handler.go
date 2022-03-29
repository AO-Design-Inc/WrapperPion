package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/pion/webrtc/v3"

	//"github.com/pion/mediadevices/pkg/codec/vpx"
	"github.com/pion/mediadevices/pkg/codec/openh264"
	//"github.com/pion/mediadevices/pkg/codec/x264"

	_ "github.com/pion/mediadevices/pkg/driver/screen"

	"encoding/json"
	"fmt"
	"sync"
)

type JSONString *C.char

var peerChannel = make(chan *webrtc.PeerConnection)
var configChannel = make(chan *webrtc.Configuration)
var sdpChannel = make(chan *webrtc.SessionDescription)
var closeChannel = make(chan struct{})
var remoteSdpChannel = make(chan *webrtc.SessionDescription)
var iceChannel = make(chan *webrtc.ICECandidateInit)
var killChannel = make(chan struct{})
var codecChannel = make(chan *mediadevices.CodecSelector)
var selectCodec = make(chan struct{})
var trackChannel = make(chan *webrtc.TrackLocal)

func addStream(codec *mediadevices.CodecSelector) *mediadevices.MediaStream {
	stream, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			constraint.FrameFormat = prop.FrameFormat(frame.FormatI420)
			constraint.FrameRate = prop.Float(120)
			constraint.Width = prop.Int(1280)
			constraint.Height = prop.Int(720)
		},
		Codec: codec,
	})
	if err != nil {
		panic(err)
	}
	return &stream
}

func getCodec() {
	h264Params, err := openh264.NewParams()
	//vp9Params, err := vpx.NewVP9Params()
	//vp8Params, err := vpx.NewVP8Params()
	//x264Params, err := x264.NewParams()
	if err != nil {
		panic(err)
	}
	h264Params.BitRate = 5_000_000
	h264Params.KeyFrameInterval = 200
	//vp8Params.BitRate = 10_000_000
	//x264Params.BitRate = 2_000_000
	//x264Params.Preset = x264.PresetVeryfast
	//vp8Params.LagInFrames = 0
	//vp8Params.KeyFrameInterval = 200
	//vp8Params.RateControlEndUsage = vpx.RateControlVBR

	codecSelector := mediadevices.NewCodecSelector(
		mediadevices.WithVideoEncoders(&h264Params),
		//mediadevices.WithVideoEncoders(&vp8Params),
		//mediadevices.WithVideoEncoders(&x264Params),
	)
	codecChannel <- codecSelector
}

func peerLifeCycle() {
	codec := <-codecChannel
	mediaEngine := webrtc.MediaEngine{}
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
	peerConnection, err := api.NewPeerConnection(*<-configChannel)
	if err != nil {
		panic(err)
	}
	var connLock sync.Mutex

	stream := *addStream(codec)
	for _, track := range stream.GetTracks() {
		fmt.Printf("%v\n", track)
		track.OnEnded(func(err error) {
			fmt.Printf("Track (ID: %s) ended with error: %v\n",
				track.ID(), err)
		})

		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			})
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := track.Close(); err != nil {
				panic(err)
			}
		}()
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState)
	})

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	go func() {
		for iceCandidate := range iceChannel {
			connLock.Lock()
			peerConnection.AddICECandidate(*iceCandidate)
			connLock.Unlock()
		}
	}()
	defer close(iceChannel)

	<-gatherComplete
	sdpChannel <- peerConnection.LocalDescription()
	remoteSdp := <-remoteSdpChannel
	connLock.Lock()
	if err := peerConnection.SetRemoteDescription(*remoteSdp); err != nil {
		panic(err)
	}
	connLock.Unlock()

	<-closeChannel
	connLock.Lock()
	if err := peerConnection.Close(); err != nil {
		panic(err)
	}
	connLock.Unlock()
}

//export AddIceCandidate
func AddIceCandidate(iceCandidateString *C.char) bool {
	var candidate webrtc.ICECandidateInit
	if err := json.Unmarshal([]byte(C.GoString(iceCandidateString)), &candidate); err != nil {
		return false
	}

	select {
	case iceChannel <- &candidate:
		return true
	default:
		return false
	}
}

//export SetRemoteDescription
func SetRemoteDescription(remoteDescString JSONString) bool {
	var desc webrtc.SessionDescription
	if err := json.Unmarshal([]byte(C.GoString(remoteDescString)), &desc); err != nil {
		return false
	}
	select {
	case remoteSdpChannel <- &desc:
		return true
	default:
		panic("wrong state")
	}
}

//export SpawnConnection
func SpawnConnection(iceValues JSONString) *C.char {
	var iceServers []webrtc.ICEServer
	if err := json.Unmarshal([]byte(C.GoString(iceValues)), &iceServers); err != nil {
		return C.CString(err.Error())
	}

	config := webrtc.Configuration{
		ICEServers: iceServers,
	}

	select {
	case configChannel <- &config:
		offerString, err := json.Marshal(<-sdpChannel)
		if err != nil {
			panic(err)
		}
		return C.CString(string(offerString))
	default:
		panic("wrong state")
	}
}

//export CloseConnection
func CloseConnection() bool {
	select {
	case closeChannel <- struct{}{}:
		return true
	default:
		panic("wrong state")
	}
}

func main() {
	go getCodec()
	peerLifeCycle()
}
