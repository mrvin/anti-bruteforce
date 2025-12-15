//go:generate protoc -I=../../api/ --go_out=../../pkg/api --go-grpc_out=require_unimplemented_servers=false:../../pkg/api ../../api/anti_bruteforce_service.proto
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/mrvin/hw-otus-go/anti-bruteforce/pkg/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

const contextTimeout = time.Second

func main() { //nolint:funlen,gocognit,cyclop
	conn, err := grpc.NewClient("localhost:55555", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("books-client: %v", err)
	}
	defer conn.Close()
	client := api.NewAntiBruteForceServiceClient(conn)
exit:
	for {
		fmt.Printf("0 - Exit\n" +
			"1 - Check request\n" +
			"2 - Add an IPv4 network address to the whitelist\n" +
			"3 - Remove an IPv4 network address from the whitelist\n" +
			"4 - Show whitelist\n" +
			"5 - Add an IPv4 network address to the blacklist\n" +
			"6 - Remove an IPv4 network address from the blacklist\n" +
			"7 - Show blacklist\n")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		switch []byte(input)[0] {
		case '0':
			fmt.Printf("Exit\n")
			break exit
		case '1':
			var req api.ReqAllowAuthorization
			fmt.Printf("Login:")
			req.Login, _ = reader.ReadString('\n')
			req.Login = strings.TrimSuffix(req.GetLogin(), "\n")
			fmt.Printf("Password:")
			req.Password, _ = reader.ReadString('\n')
			req.Password = strings.TrimSuffix(req.GetPassword(), "\n")
			fmt.Printf("IP:")
			req.Ip, _ = reader.ReadString('\n')
			req.Ip = strings.TrimSuffix(req.GetIp(), "\n")

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()
			res, err := client.AllowAuthorization(ctx, &req)
			if err != nil {
				log.Printf("Check request: %v", err)
				continue
			}
			fmt.Printf("result: %t\n", res.GetAllow())
		case '2':
			var req api.ReqNetwork
			fmt.Printf("Network:")
			req.Network, _ = reader.ReadString('\n')
			req.Network = strings.TrimSuffix(req.GetNetwork(), "\n")

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()

			_, err := client.AddNetworkToWhitelist(ctx, &req)
			if err != nil {
				log.Printf("Add network to whitelist: %v", err)
				continue
			}
			fmt.Println("Success")
		case '3':
			var req api.ReqNetwork
			fmt.Printf("Network:")
			req.Network, _ = reader.ReadString('\n')
			req.Network = strings.TrimSuffix(req.GetNetwork(), "\n")

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()

			_, err := client.DeleteNetworkFromWhitelist(ctx, &req)
			if err != nil {
				log.Printf("Remove network from whitelist: %v", err)
				continue
			}
			fmt.Println("Success")
		case '4':
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()
			res, err := client.Whitelist(ctx, &emptypb.Empty{})
			if err != nil {
				log.Printf("Whitelist: %v", err)
				continue
			}
			fmt.Println("Whitelist:")
			for _, network := range res.GetNetworks() {
				fmt.Println(network)
			}
		case '5':
			var req api.ReqNetwork
			fmt.Printf("Network:")
			req.Network, _ = reader.ReadString('\n')
			req.Network = strings.TrimSuffix(req.GetNetwork(), "\n")

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()

			_, err := client.AddNetworkToBlacklist(ctx, &req)
			if err != nil {
				log.Printf("Add network to blacklist: %v", err)
				continue
			}
			fmt.Println("Success")
		case '6':
			var req api.ReqNetwork
			fmt.Printf("Network:")
			req.Network, _ = reader.ReadString('\n')
			req.Network = strings.TrimSuffix(req.GetNetwork(), "\n")

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()

			_, err := client.DeleteNetworkFromBlacklist(ctx, &req)
			if err != nil {
				log.Printf("Remove network from blacklist: %v", err)
				continue
			}
			fmt.Println("Success")
		case '7':
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			defer cancel()
			res, err := client.Blacklist(ctx, &emptypb.Empty{})
			if err != nil {
				log.Printf("Blacklist: %v", err)
				continue
			}
			fmt.Println("Blacklist:")
			for _, network := range res.GetNetworks() {
				fmt.Println(network)
			}
		default:
		}
	}
}
