package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/datakit-dev/dtkt-sdk/sdk-go/common"
)

func main() {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		var (
			outFile  string
			outCount int
		)
		if len(os.Args) > 1 {
			outFile = os.Args[1]
		}

		player, err := NewPlayer()
		if err != nil {
			panic(err)
		}

		// Ensure player is closed on exit
		defer player.Close()
		player.Play()

		// Set up signal handling for graceful shutdown
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-sigs
			fmt.Println("\nReceived interrupt, shutting down.")
			os.Exit(0)
		}()

		// Increase scanner buffer size if needed
		scanner := bufio.NewScanner(os.Stdin)
		buf := make([]byte, 0, 64*1024) // 64 KB buffer
		scanner.Buffer(buf, 1024*1024)  // 1 MB max buffer

		for scanner.Scan() {
			eventMap, err := common.UnmarshalJSON[common.JSONMap](scanner.Text())
			if err != nil {
				fmt.Println("Error unmarshalling event to map:", err.Error())
				continue
			}

			eventType, ok := common.GetJSONValue[string](eventMap, ".type")
			if !ok {
				fmt.Println("Error extracting event type:", eventType)
				continue
			}

			fmt.Println("Event", eventType)

			switch eventType {
			case "response.audio.delta":
				audioBase64, ok := common.GetJSONValue[string](eventMap, ".delta")
				if !ok {
					continue
				}

				audioBytes, err := base64.StdEncoding.DecodeString(audioBase64)
				if err != nil {
					fmt.Println("Error decoding audio base64:", err)
					continue
				}

				player.AddChunk(audioBytes)

				if outFile != "" {
					file := strings.Replace(outFile, ".pcm", fmt.Sprintf("-%d.pcm", outCount), 1)

					// Append audio chunk to file
					out, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err != nil {
						fmt.Println("Error opening audio file:", err.Error())
						continue
					}
					defer func() {
						if err := out.Close(); err != nil {
							fmt.Println("Error closing audio file:", err.Error())
						}
					}()

					fmt.Println("Writing audio chunk:", file)
					_, err = out.Write(audioBytes)
					if err != nil {
						fmt.Println("Error writing audio file:", err.Error())
					}
				}
			case "response.audio.done":
				outCount++
			default:
				// transcript, ok := common.JSONValue[string](eventMap, ".item.content[0].transcript")
				// if !ok {
				// 	continue
				// }
				fmt.Println(scanner.Text())
			}
		}

		// Check for errors in scanner
		if err := scanner.Err(); err != nil {
			fmt.Println("Error reading stdin:", err)
		}
	} else {
		panic(fmt.Errorf("stdin is empty"))
	}
}
