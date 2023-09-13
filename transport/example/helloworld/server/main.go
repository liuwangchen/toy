package main

import (
	"context"
	"fmt"
	"log"

	"github.com/liuwangchen/toy/app"
	"github.com/liuwangchen/toy/transport/example/helloworld/pb"
	"github.com/liuwangchen/toy/transport/middleware/recovery"
	"github.com/liuwangchen/toy/transport/rpc/grpc"
	"github.com/liuwangchen/toy/transport/rpc/httprpc"
	"github.com/liuwangchen/toy/transport/rpc/kafkarpc"
	"github.com/liuwangchen/toy/transport/rpc/natsrpc"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "helloworld"
	// Version is the version of the compiled software.
	// Version = "v1.0.0"
)

// server is used to implement helloworld.GreeterServer.
type Server struct {
	pb.UnimplementedGreeterServer
}

func (s *Server) MultiSayHello(request *pb.HelloRequest, stream pb.Greeter_MultiSayHelloServer) error {
	for i := 0; i < 10; i++ {
		err := stream.Send(&pb.HelloReply{
			Message: fmt.Sprintf("%d", i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) MultiSayHelloStream(ctx context.Context, request *pb.HelloRequest, stream *pb.MultiSayHelloServerStream) error {
	for i := 0; i < 10; i++ {
		err := stream.Send(&pb.HelloReply{
			Message: fmt.Sprintf("%d", i),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// SayHello implements helloworld.GreeterServer
func (s *Server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	if in.Name == "error" {
		return nil, fmt.Errorf("invalid argument %s", in.Name)
	}
	if in.Name == "panic" {
		panic("server panic")
	}
	return &pb.HelloReply{Message: fmt.Sprintf("Hello %+v", in.Name)}, nil
}

// server is used to implement helloworld.GreeterServer.
type PushServer struct {
}

func (p *PushServer) Push(ctx context.Context, notify *pb.PushNotify) error {
	fmt.Println("kafka receive notify ", notify.Name)
	return nil
}

func httpRunner() app.Runner {
	conn, err := httprpc.NewServerConn(
		httprpc.WithAddress(":8000"),
		httprpc.WithConnMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	pb.RegisterGreeterHTTPServer(conn, &Server{})
	return conn
}

func grpcRunner() app.Runner {
	grpcSrv, err := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	pb.RegisterGreeterServer(grpcSrv, &Server{})
	return grpcSrv
}

func natsRunner() app.Runner {
	conn, err := natsrpc.NewServerConn(
		natsrpc.WithAddr("nats://nats:wg1q2w3e@192.168.1.153:4333,nats://nats:wg1q2w3e@192.168.1.154:4333,nats://nats:wg1q2w3e@192.168.1.155:4333"),
		natsrpc.WithConnMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	err = pb.RegisterGreeterNatsServer(conn, &Server{})
	if err != nil {
		panic(err)
	}
	err = pb.RegisterPusherNatsServer(conn, &PushServer{})
	if err != nil {
		panic(err)
	}
	return conn
}

func kafkaRunner() app.Runner {
	kafkaSrv, err := kafkarpc.NewServer(
		kafkarpc.Address(":9092"),
		kafkarpc.Middleware(recovery.Recovery()),
	)
	if err != nil {
		panic(err)
	}
	pb.RegisterPusherKafkaServer(kafkaSrv, &PushServer{})
	return kafkaSrv
}

func main() {
	app := app.New(
		app.WithName(Name),
		app.WithRunners(
			//httpRunner(),
			//grpcRunner(),
			natsRunner(),
			//udpRunner(),
			//kafkaRunner(),
		),
	)

	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
