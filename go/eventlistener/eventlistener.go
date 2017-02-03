/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric/events/consumer"
	pb "github.com/hyperledger/fabric/protos"
	"encoding/json"
	"time"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/aws"
)


type adapter struct {
	notfy              chan *pb.Event_Block
	rejected           chan *pb.Event_Rejection
	cEvent             chan *pb.Event_ChaincodeEvent
	listenToRejections bool
	chaincodeID        string
}

type NetVoteEvent struct {
	VoteEvent VoteEvent
	AccountId string
	ChaincodeId string
	EventName string
	TxId string
	Timestamp int64
}

type Ballot struct{
	Id string
	Name string
	Private bool
}

type Option struct {
	Id string
	Name string
	Attributes map[string]string
}

type Decision struct {
	Id                string
	Name              string
	BallotId          string
	Options           []Option
	Attributes map[string]string
	ResponsesRequired int
	RepeatVoteDelaySeconds int
	Repeatable        bool
}

type VoterDecision struct {
	DecisionId string
	Selections map[string]int
	Reasons map[string]map[string]string
	Attributes map[string]string
}

type BallotDecisions struct{
	Ballot Ballot
	Decisions []Decision
}

//must match structure in vote.go...this is for marshalling
type VoteEvent struct {
	Ballot BallotDecisions
	Dimensions []string
	VoterAttributes map[string]string
	VoteDecisions []VoterDecision
	AccountId string
	Timestamp int64
}

//GetInterestedEvents implements consumer.EventAdapter interface for registering interested events
func (a *adapter) GetInterestedEvents() ([]*pb.Interest, error) {
	if a.chaincodeID != "" {
		return []*pb.Interest{
			{EventType: pb.EventType_CHAINCODE,
				RegInfo: &pb.Interest_ChaincodeRegInfo{
					ChaincodeRegInfo: &pb.ChaincodeReg{
						ChaincodeID: a.chaincodeID,
						EventName:   ""}}},{EventType: pb.EventType_BLOCK}}, nil
	}
	return []*pb.Interest{{EventType: pb.EventType_BLOCK}, {EventType: pb.EventType_REJECTION}}, nil
}

//Recv implements consumer.EventAdapter interface for receiving events
func (a *adapter) Recv(msg *pb.Event) (bool, error) {
	if o, e := msg.Event.(*pb.Event_Block); e {
		a.notfy <- o
		return true, nil
	}
	if o, e := msg.Event.(*pb.Event_Rejection); e {
		if a.listenToRejections {
			a.rejected <- o
		}
		return true, nil
	}
	if o, e := msg.Event.(*pb.Event_ChaincodeEvent); e {
		a.cEvent <- o
		return true, nil
	}
	return false, fmt.Errorf("Receive unkown type event: %v", msg)
}

//Disconnected implements consumer.EventAdapter interface for disconnecting
func (a *adapter) Disconnected(err error) {
	fmt.Printf("Disconnected...exiting\n")
	os.Exit(1)
}

func createEventClient(eventAddress string, listenToRejections bool, cid string) *adapter {
	var obcEHClient *consumer.EventsClient

	done := make(chan *pb.Event_Block)
	reject := make(chan *pb.Event_Rejection)
	adapter := &adapter{notfy: done, rejected: reject, listenToRejections: listenToRejections, chaincodeID: cid, cEvent: make(chan *pb.Event_ChaincodeEvent)}
	obcEHClient, _ = consumer.NewEventsClient(eventAddress, 5, adapter)
	if err := obcEHClient.Start(); err != nil {
		fmt.Printf("could not start chat %s\n", err)
		obcEHClient.Stop()
		return nil
	}

	return adapter
}

func main() {
	chaincodeID, cidExists := os.LookupEnv("CHAINCODE_ID")
	eventAddress, eaExists := os.LookupEnv("PEER_HOST")
	streamName, snExists := os.LookupEnv("STREAM_NAME")

	if(!cidExists){
		fmt.Println("CHAINCODE_ID env variable is not set")
		os.Exit(1)
	}
	if(!eaExists){
		fmt.Println("PEER_HOST env variable is not set")
		os.Exit(1)
	}

	if(!snExists){
		fmt.Println("STREAM_NAME env variable is not set, defaulting to votes")
		streamName = "votes"
	}

	sess, err := session.NewSession()
	if err != nil {
		fmt.Println("failed to create session,", err)
		return
	}

	svc := kinesis.New(sess)

	fmt.Printf("Event Address: %s\n", eventAddress)
	fmt.Printf("ChaincodeID: %s\n", chaincodeID)

	a := createEventClient(eventAddress, false, chaincodeID)
	if a == nil {
		fmt.Println("Error creating event client")
		return
	}

	for {
		select {
		case b := <-a.notfy:
			fmt.Printf("\n")
			fmt.Printf("\n")
			fmt.Printf("Received block\n")
			fmt.Printf("--------------\n")
			for _, r := range b.Block.Transactions {
				fmt.Printf("Transaction:\n\t[%v]\n", r)
			}
			fmt.Printf("Transaction:\n\t[%v]\n", string(b.Block.StateHash))
		case ce := <-a.cEvent:

			fmt.Printf("\nEVENT BYTES:\n"+string(ce.ChaincodeEvent.Payload)+"\n")


			var evt map[string]interface{}
			json.Unmarshal(ce.ChaincodeEvent.Payload, &evt)

			var vote VoteEvent
			json.Unmarshal(ce.ChaincodeEvent.Payload, &vote)

			nowTime := time.Now().UnixNano()
			vote.Timestamp = nowTime

			netvoteEvent := NetVoteEvent{
				AccountId: evt["AccountId"].(string),
				VoteEvent: vote,
				TxId:ce.ChaincodeEvent.TxID,
				Timestamp: nowTime,
				EventName: ce.ChaincodeEvent.EventName,
				ChaincodeId: ce.ChaincodeEvent.ChaincodeID}

			eventBytes, _ := json.Marshal(netvoteEvent)

			fmt.Printf("\nNETVOTE EVENT:\n"+string(eventBytes)+"\n")

			params := &kinesis.PutRecordInput{
				Data:                      eventBytes,
				PartitionKey:              aws.String(evt["AccountId"].(string)),
				StreamName:                aws.String(streamName),
			}
			resp, err := svc.PutRecord(params)

			if err != nil {
				fmt.Println(err.Error())
			}

			fmt.Println(resp)
		}
	}

}
