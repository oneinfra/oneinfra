package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	criapi "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
)

const (
	address = "passthrough:///unix:///var/run/docker/containerd/containerd.sock"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	criClient := criapi.NewRuntimeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	response, err := criClient.Version(ctx, &criapi.VersionRequest{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(response)
}
