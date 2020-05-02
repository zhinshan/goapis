package main

import (
	"apis/apipb"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
)

func main() {
	log.Println("Welcome to ApisClient!")

	cc, err := grpc.Dial("localhost:8199", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Dial error=%v", err)
	}
	defer cc.Close()

	client := apipb.NewAipsServiceClient(cc)
	//log.Printf("The client=%v", client)

	// callUnariapi(client)
	// callClientStream(client)
	// callServerStream(client, 9)
	//callBidirection(client)
	// callUnaryWidthDeadline(client, 1*time.Second)
	// callUnaryWidthDeadline(client, 5*time.Second)
	callErrorHandle(client)

}

func callUnariapi(c apipb.AipsServiceClient) {
	log.Println("Call Unari() ...")

	resp, err := c.Unary(context.Background(),
		&apipb.UnaryRequest{
			ReqMsg: "aloha",
		})
	if err != nil {
		log.Fatalf("Server error=%v", err)
	}

	log.Printf("Server response msg=%v", resp.GetRespMsg())
}

func callClientStream(c apipb.AipsServiceClient) {
	log.Println("Call ClientStream() ...")

	s, err := c.ClientStream(context.Background())
	if err != nil {
		log.Fatalf("ClientStream error=%v", err)
	}

	listReq := []apipb.ClientStreamRequest{
		{Num: 11},
		{Num: 52},
		{Num: 32},
		{Num: 94},
		{Num: 45},
		{Num: 2},
		{Num: 24},
		{Num: 5},
	}

	for _, req := range listReq {
		log.Printf("ClientStream: send to server number=%v", req.GetNum())
		err := s.Send(&req)
		if err != nil {
			log.Fatalf("ClientStream: send to server error=%v", err)
		}

		time.Sleep(1 * time.Second)
	}

	resp, err := s.CloseAndRecv()
	if err != nil {
		log.Printf("ClientStream: closed and received error=%v", err)
	}

	log.Printf("ClientStream: Received from server max=%v, min=%v, average=%v", resp.GetMax(), resp.GetMin(), resp.GetAverage())
}

func callServerStream(c apipb.AipsServiceClient, fibonum int32) {
	log.Println("Call ServerStream() ...")

	s, err := c.ServerStream(context.Background(), &apipb.ServerStreamRequest{
		Fibonum: fibonum,
	})

	if err != nil {
		log.Fatalf("ServerStream error=%v", err)
	}

	for {
		resp, err := s.Recv()
		if err != nil {
			log.Printf("ServerStream: Received error=%v", err)
			break
		}

		log.Printf("ServerStream: Fibonacy(%v) = %v", resp.GetFibopos(), resp.GetFibonacy())
	}

	log.Println("ServerStream: Finished.")
}

func callBidirection(c apipb.AipsServiceClient) {
	log.Println("Call Bidirection() ...")

	s, err := c.Bidirection(context.Background())
	if err != nil {
		log.Fatalf("Bidirection error=%v", err)
	}

	waitc := make(chan struct{})

	// routine of client streaming-send to server task
	go func() {
		listReq := []apipb.BidirectionRequest{
			{Num: 11},
			{Num: 52},
			{Num: 32},
			{Num: 94},
			{Num: 45},
			{Num: 2},
			{Num: 24},
			{Num: 5},
		}

		for _, req := range listReq {
			log.Printf("Bidirection: send to server number=%v", req.GetNum())
			err := s.Send(&req)
			if err != nil {
				log.Fatalf("Bidirection: send to server error=%v", err)
			}

			time.Sleep(100 * time.Millisecond)
		}

		s.CloseSend()

	}()

	// routine of client stream-recived from server task
	go func() {
		for {
			resp, err := s.Recv()
			if err != nil {
				log.Printf("Bidirection: Received error=%v", err)
				break
			}

			log.Printf("Bidirection: max=%v, min=%v", resp.GetMax(), resp.GetMin())
		}

		close(waitc)

	}()

	<-waitc
}

func callUnaryWidthDeadline(c apipb.AipsServiceClient, timeout time.Duration) {
	log.Println("Call UnaryWidthDeadline() ...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	resp, err := c.UnaryWidthDeadline(ctx,
		&apipb.UnaryRequest{
			ReqMsg: "aloha",
		})
	if err != nil {
		log.Printf("Server error=%v, msg=%v", err, resp.GetRespMsg())
		return
	}

	log.Printf("Server response msg=%v", resp.GetRespMsg())
}

func callErrorHandle(c apipb.AipsServiceClient) {
	log.Println("Call ErrorHandle() ...")

	resp, err := c.ErrorHandle(context.Background(), &apipb.UnaryRequest{})
	if err != nil {
		log.Printf("ErrorHandle: Error=%v", err)
		return
	}

	log.Printf("ErrorHandle: Response=%v", resp)
}
