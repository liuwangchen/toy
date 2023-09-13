package main

import (
	"context"
	"fmt"
	"log"

	"github.com/liuwangchen/toy/transport/example/helloworld/pb"
	"github.com/liuwangchen/toy/transport/middleware/recovery"
	conngrpc "github.com/liuwangchen/toy/transport/rpc/grpc"
	"github.com/liuwangchen/toy/transport/rpc/httprpc"
	"github.com/liuwangchen/toy/transport/rpc/kafkarpc"
	"github.com/liuwangchen/toy/transport/rpc/natsrpc"
)

func main() {
	//callHTTP()
	//callKafka()
	//callUdp()
	//callGRPC()
	callNats()
}

func callHTTP() {
	conn, err := httprpc.NewClientConn(
		context.Background(),
		httprpc.WithConnMiddleware(
			recovery.Recovery(),
		),
		httprpc.WithAddress("127.0.0.1:8000"),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewGreeterHTTPClient(conn)
	stream, err := client.MultiSayHelloStream(context.Background(), &pb.HelloRequest{Name: "goctopus"})
	if err != nil {
		log.Fatal(err)
	}
	for {
		rep, err := stream.Recv()
		if err == nil {
			log.Printf("[http] SayHello %s\n", rep.Message)
		}
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func callGRPC() {
	conn, err := conngrpc.DialInsecure(
		context.Background(),
		conngrpc.WithEndpoint("127.0.0.1:9000"),
		conngrpc.WithMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "goctopus"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[grpc] SayHello %+v\n", reply)

	// returns error
	_, err = client.SayHello(context.Background(), &pb.HelloRequest{Name: "error"})
	log.Printf("[grpc] SayHello error: %v\n", err)
}

func callNats() {
	conn, err := natsrpc.NewClientConn(
		natsrpc.WithAddr("nats://nats:wg1q2w3e@192.168.1.153:4333,nats://nats:wg1q2w3e@192.168.1.154:4333,nats://nats:wg1q2w3e@192.168.1.155:4333"),
		natsrpc.WithConnMiddleware(
			recovery.Recovery(),
		),
	)
	if err != nil {
		panic(err)
	}
	client := pb.NewGreeterNatsClient(conn)
	reply, err := client.SayHello(context.Background(), &pb.HelloRequest{Name: "goctopus"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[nats] SayHello %+v\n", reply)

	// returns error
	_, err = client.SayHello(context.Background(), &pb.HelloRequest{Name: "error"})
	log.Printf("[nats] SayHello error: %v\n", err)
}

func callKafka() {
	conn, err := kafkarpc.NewClient(
		context.Background(),
		kafkarpc.WithEndpoint(":9092"),
	)
	if err != nil {
		panic(err)
	}
	client := pb.NewPusherKafkaClient(conn)
	err = client.Push(context.Background(), &pb.PushNotify{Name: "goctopus"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("[kafka] push\n")
}
