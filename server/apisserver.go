package main

import (
	"apis/apipb"
	"context"
	"io"
	"log"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct{}

func (*server) Unary(ctx context.Context, in *apipb.UnaryRequest) (*apipb.UnaryResponse, error) {
	log.Printf("Invoke Unary() ... msg=%v", in.GetReqMsg())

	resp := &apipb.UnaryResponse{
		RespMsg: "Hello " + in.GetReqMsg(),
	}

	return resp, nil
}

func (*server) ClientStream(s apipb.AipsService_ClientStreamServer) error {
	log.Printf("Invoke ClientStream() ...")

	var total int32
	var count int32
	resp := &apipb.ClientStreamResponse{
		Max:     0,
		Min:     20000,
		Average: 10000,
	}

	for {
		req, err := s.Recv()
		if err == io.EOF {
			log.Println("ClientStream: no more from client, send and close.")
			resp.Average = float32(total / count)
			return s.SendAndClose(resp)
		}

		if err != nil {
			log.Printf("ClientStream: read from stream error=%v", err)
			return err
		}

		num := req.GetNum()
		log.Printf("ClientStream: read number=%v", num)
		count++
		total += num
		if resp.Max < num {
			resp.Max = num
		}
		if resp.Min > num {
			resp.Min = num
		}
	}
}

func (*server) ServerStream(req *apipb.ServerStreamRequest, s apipb.AipsService_ServerStreamServer) error {
	log.Printf("Invoke ServerStream() ... fibonacy number of step=%v", req.GetFibonum())

	fibo1 := int32(1)
	fibo2 := int32(1)

	// f(n) = 1; n=1
	// f(n) = 1; n=2
	// f(n) = f(n-1) + f(n-2); n>2
	for i := 1; i <= int(req.GetFibonum()); i++ {
		log.Printf("ServerStream: Fibonacy(%v) = %v", i, fibo1)
		err := s.Send(&apipb.ServerStreamResponse{
			Fibonacy: fibo1,
			Fibopos:  int32(i),
		})

		if err != nil {
			log.Printf("ServerStream: send to client error=%v", err)
			return nil
		}

		fibonext := fibo1 + fibo2
		fibo1 = fibo2
		fibo2 = fibonext

		time.Sleep(200 * time.Millisecond)
	}

	return nil
}

func (*server) Bidirection(s apipb.AipsService_BidirectionServer) error {
	log.Printf("Invoke Bidirection() ...")

	resp := &apipb.BidirectionResponse{
		Max: 0,
		Min: 20000,
	}

	// server had one-double task, 1: Received from client, and, 2: Send to client

	for {
		req, err := s.Recv()
		if err == io.EOF {
			log.Println("Bidirection: no more from client, close.")
			return nil
		}

		if err != nil {
			log.Printf("Bidirection: read from stream error=%v", err)
			return err
		}

		num := req.GetNum()
		log.Printf("Bidirection: read number=%v", num)
		if resp.Max < num {
			resp.Max = num
		}
		if resp.Min > num {
			resp.Min = num
		}

		err = s.Send(resp)
		if err != nil {
			log.Printf("Bidirection: Send to client error=%v", err)
			return nil
		}
	}
}

func (*server) UnaryWidthDeadline(ctx context.Context, in *apipb.UnaryRequest) (*apipb.UnaryResponse, error) {
	log.Println("Server: call UnaryWidthDeadline() ...")

	for i := 0; i < 3; i++ {
		log.Printf("UnaryWidthDeadline: context error=%v", ctx.Err())
		if ctx.Err() == context.Canceled {
			return &apipb.UnaryResponse{
				RespMsg: "Server response to client none-normaly!",
			}, nil
		}

		time.Sleep(1 * time.Second)
	}

	return &apipb.UnaryResponse{
		RespMsg: "Server response to client normaly!",
	}, nil
}

func (*server) ErrorHandle(ctx context.Context, in *apipb.UnaryRequest) (*apipb.UnaryResponse, error) {
	log.Println("Server: call ErrorHandle() ...")

	return nil, status.Errorf(codes.FailedPrecondition, "Server cancel request!")
}

func main() {
	log.Println("Welcome to ApisServer!")

	lis, err := net.Listen("tcp", "localhost:8199")
	if err != nil {
		log.Fatalf("Listen error=%v", err)
	}

	s := grpc.NewServer()
	apipb.RegisterAipsServiceServer(s, &server{})

	err = s.Serve(lis)
	if err != nil {
		log.Fatalf("Serve the service error=%v", err)
	}
}
