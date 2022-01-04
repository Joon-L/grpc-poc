package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "echo/proto"
	"net"
	"net/http"
)

type server struct{}

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})
}

type healthHandler struct {
}

func (m *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logrus.Println("Listening on 5001: health ")
	w.Write([]byte("Listening on 5001: foo "))
}

func main() {
	go func() {
		http.Handle("/healthz", &healthHandler{})
		http.ListenAndServe(":5001", nil)
	}()

	listener, err := net.Listen("tcp", ":5000")
	if err != nil {
		logrus.Fatalf("Listener exception %v", err.Error())
	}

	grpcServer := grpc.NewServer()
	pb.RegisterScorerServer(grpcServer, &scorerServer{})
	if err = grpcServer.Serve(listener); err != nil {
		logrus.Fatalf("While serving gRpc request: %v", err)
	}
}

type scorerServer struct {
	pb.UnimplementedScorerServer
}

func (s *scorerServer) Score(ctx context.Context, request *pb.InferenceRequest) (*pb.InferenceResponse, error) {
	logrus.Printf("Unary Received: %v", request.GetPrompt())
	return &pb.InferenceResponse{
		Result: request.GetPrompt() + " sunny",
	}, nil
}

func (s *scorerServer) StreamingRequestScore(stream pb.Scorer_StreamingRequestScoreServer) error {
	result := []string{"START "}
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			finalResult := strings.Join(result, "__") + " END"
			logrus.Printf("cStream End of streaming request, will return the response %s", finalResult)
			return stream.SendAndClose(&pb.InferenceResponse{
				Result: finalResult,
			})
		}
		if err != nil {
			return nil
		}
		if len(result) == 1 {
			logrus.Println("cStream First response received from client")
		}
		result = append(result, request.GetPrompt())
	}
}

func (s *scorerServer) StreamingResponseScore(request *pb.InferenceRequest, stream pb.Scorer_StreamingResponseScoreServer) error {
	prompt := request.GetPrompt()
	logrus.Println("sStream Sending first response for the Server Streaming request")
	for i := 0; i < 10; i++ {
		result := fmt.Sprintf("%s %v", prompt, i)
		error := stream.Send(&pb.InferenceResponse{
			Result: result,
		})
		if error != nil {
			logrus.Printf("Error in processing Server streaming request %v", error)
		}
		time.Sleep(250 * time.Millisecond)
	}
	logrus.Println("sStream Sent all the request to client")
	return nil
}

func (s *scorerServer) BidirectionalScore(stream pb.Scorer_BidirectionalScoreServer) error {
	logrus.Println("BiDi Starting the bidirectional request processing")
	result := []string{"BATCH START "}
	for i := 0; i < 10; i++ {
		request, error := stream.Recv()
		if error != nil {
			logrus.Printf("Could not process bidirection request %v", error)
			return error
		} else {
			result = append(result, request.GetPrompt())
		}

		if i%2 == 0 {
			batchResult := strings.Join(result, "__") + " BATCH END"
			logrus.Printf("BiDi Sending current stream response %s", batchResult)
			stream.Send(&pb.InferenceResponse{
				Result: batchResult,
			})
			result = []string{"BATCH START "}
		}
		time.Sleep(125 * time.Millisecond)
	}
	logrus.Println("BiDi Ending the bidirectional request processing")
	return nil
}
