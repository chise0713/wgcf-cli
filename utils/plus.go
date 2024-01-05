package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// UGLY CODE , Pull Request is welcome
// This may cause some memery leak, DO NOT use for a long time.
func Plus(filePath string, i int) error {
	var times sync.WaitGroup
	var err error
	var id string
	var currentStep int = i
	ctx, cancel := context.WithCancel(context.Background())
	C, Ca := context.WithCancel(context.Background())
	if _, id, err = GetTokenID(filePath); err != nil {
		panic(err)
	}
	go func() {
		for {
			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
			select {
			case <-C.Done():
				return
			case <-signalCh:
				cancel()
				fmt.Println("\nWaiting for Response...")
				go func() {
					signalCh := make(chan os.Signal, 1)
					signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
					<-signalCh
					fmt.Println()
					os.Exit(1)
				}()
				times.Wait()
				fmt.Println("Updating config file...")
				UpdateConfigFile(filePath)
				fmt.Println("Updated config file successfully")
				os.Exit(0)
			}

		}
	}()

	for {
		time.Sleep(500 * time.Millisecond)
		times.Add(1)
		go func(index int) {
			var publicKey string
			defer times.Done()
			select {
			case <-ctx.Done():
				return
			default:
			}
			currentStep++
			if _, publicKey, err = GenerateKey(); err != nil {
				panic(err)
			}

			installID := RandStringRunes(22, nil)
			fcmtoken := RandStringRunes(134, nil)

			payload := []byte(
				`{
					"key":"` + publicKey + `",
					"install_id":"` + installID + `",
					"fcm_token":"` + installID + `:APA91b` + fcmtoken + `",
					"tos":"` + time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + `",
					"model":"Android",
					"referrer": "` + id + `",
					"serial_number":"` + installID + `"
				}`,
			)
			fmt.Println("Registering...", index, "times")
			if _, err = request(payload, "", "", "register"); err != nil {
				fmt.Println(err)
				fmt.Println("Waiting for 30 seconds...")
				select {
				case <-ctx.Done():
					return
				default:
				}
				cancel()
				time.Sleep(30 * time.Second)

				Ca()
				Plus(filePath, i)
			}
			i++
			fmt.Println("						Added", index, "GB")
		}(currentStep)
	}
}
