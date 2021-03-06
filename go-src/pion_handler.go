// a handler for pion
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
)

type JSONString *C.char
//var peerConnection *webrtc.PeerConnection
var pc_channel = make(chan *webrtc.PeerConnection, 1)
var connectionLock = make(chan struct{}, 1)


func peerConnector(config *webrtc.Configuration, recvSdp chan *C.char) {

    
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

	mediaEngine := webrtc.MediaEngine{}
	codecSelector.Populate(&mediaEngine)
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&mediaEngine))
  peerConnection, err := api.NewPeerConnection(*config)
  pc_channel <- peerConnection
	if err != nil {
		panic(err)
	}

	stream, err := mediadevices.GetDisplayMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			constraint.FrameFormat = prop.FrameFormat(frame.FormatI420)
			constraint.FrameRate = prop.Float(120)
      constraint.Width = prop.Int(1280)
      constraint.Height = prop.Int(720)
		},
		Codec: codecSelector,
	})

	for _, track := range stream.GetTracks() {
    fmt.Printf("%v\n", track)
		track.OnEnded(func(err error) {
			fmt.Printf("Track (ID: %s) ended with error: %v\n",
				track.ID(), err)
		})

		_, err = peerConnection.AddTransceiverFromTrack(track,
			webrtc.RtpTransceiverInit{
				Direction: webrtc.RTPTransceiverDirectionSendonly,
			},
		)
		if err != nil {
			panic(err)
		}
	}
	if err != nil {
		panic(err)
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s \n", connectionState)
    if connectionState == webrtc.ICEConnectionStateFailed ||
       connectionState == webrtc.ICEConnectionStateClosed {
          fmt.Println("now it is time to die");
          <-connectionLock
        }
	})
	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		panic(err)
	}

	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	if err = peerConnection.SetLocalDescription(offer); err != nil {
		panic(err)
	}

	<-gatherComplete
	offerString, err := json.Marshal(*peerConnection.LocalDescription())
	cOfferString := C.CString(string(offerString))
	recvSdp <- cOfferString

  //HOLD until close
  connectionLock <- struct{}{}
  <-connectionLock
}

//export SpawnConnection
func SpawnConnection(iceValues JSONString) *C.char {
  //note if this returns an empty string its waiting
  select {
    case connectionLock<-struct{}{}:
      goto cont
    default:
      return C.CString("")
    }
  cont:
  fmt.Println("lucrative")
	sdpRecv := make(chan *C.char, 1)
	var iceServers []webrtc.ICEServer
	if err := json.Unmarshal([]byte(C.GoString(iceValues)), &iceServers); err != nil {
    return C.CString(err.Error())
	}


	config := webrtc.Configuration{
		ICEServers: iceServers,
	}

	go peerConnector(&config, sdpRecv)

	
	return(<-sdpRecv)
}

//export SetRemoteDescription
func SetRemoteDescription(remoteDescString JSONString) bool {
	var desc webrtc.SessionDescription
	if err := json.Unmarshal([]byte(C.GoString(remoteDescString)), &desc); err != nil {
		return false
	}
  //go remoteSetter(&desc)
  peerConnection := <-pc_channel
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete
	if err := peerConnection.SetRemoteDescription(desc); err != nil {
		panic(err)
	}
  pc_channel <- peerConnection
	return true
}

//export AddIceCandidate
func AddIceCandidate(iceCandidateString *C.char) bool {
  peerConnection := <-pc_channel
  var candidate webrtc.ICECandidateInit
  if err := json.Unmarshal([]byte(C.GoString(iceCandidateString)), &candidate); err != nil {
    return false
  }

  if err := peerConnection.AddICECandidate(candidate); err != nil {
    panic(err)
  }

  pc_channel <- peerConnection
  return true
}

//export CloseConnection
func CloseConnection() bool {
  peerConnection := <-pc_channel 
  if err := peerConnection.Close(); err != nil {
    panic(err)
  }

  return true
}

  




/*
func remoteSetter(desc *webrtc.SessionDescription) {
	gatherComplete := webrtc.GatheringCompletePromise(peerConnection)
	<-gatherComplete
	if err := peerConnection.SetRemoteDescription(*desc); err != nil {
		panic(err)
	}
}
*/


func main() {}
