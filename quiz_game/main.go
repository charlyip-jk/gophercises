package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"bufio"
	"math/rand"
)

func main() {

	// adding flags
	filePath := flag.String("filePath", "./problems.csv", "Specify the file path. Defaults to problems.csv.")
	timeOut := flag.Int("timeOut", 30, "Specify the question timeout. Defaults to 30 seconds.")
	shuffle := flag.Bool("shuffle", false, "Should the questions be shuffled? Defaults to false.")
	flag.Parse()

	// retrieving questions
	questions := readCsvFile(*filePath)
	if *shuffle {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(questions), func(i, j int) {
			questions[i], questions[j] = questions[j], questions[i]
		})
	}

	correctAnswers := startQuiz(questions, *timeOut)
	if correctAnswers != nil {
		fmt.Println(*correctAnswers, "questions answered correctly out of", len(questions))
	} else {
		fmt.Println("Didn't finish the quiz in time! Please try again or set a longer timer.")
	}
}

func startQuiz(questions [][]string, timer int) *int {

	// wait for enter or key-press to start timer
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please press enter to start the quiz.")
	reader.ReadString('\n')

	// Create a channel to receive a signal when the timer is done or interrupted
	doneTimer := make(chan struct{})

	// Create a channel to receive signal when quiz was finished before timeout
	resultCh := make(chan int)

	// Start the countdown timer in the background
	go startCountdown(timer, doneTimer) // Replace 10 with the desired countdown time in seconds

	// quiz logic
	go func() {
		counter := 0
		for _, entry := range questions {

			fmt.Println(entry[0], "?")
			fmt.Print("Enter text: ")
			text, _ := reader.ReadString('\n')

			if strings.TrimRight(text, "\n") == entry[1] {
				counter++
			}
		}
		resultCh <- counter // Send the result through the channel
	}()

	// Wait for the timer to complete or for an interrupt signal
	select {
	case counter := <-resultCh:
		// Quiz completed before timeout, return the counter
		fmt.Println("Quiz completed.")
		return &counter
	case <-doneTimer:
		// Countdown completed, do any cleanup or final tasks here
		fmt.Println("Countdown completed.")
		return nil
	case <-interruptSignal():
		// Program interrupted by user, quit the program
		fmt.Println("Program interrupted. Exiting...")
		return nil
	}
}

func readCsvFile(filePath string) [][]string {
	
    f, err := os.Open(filePath)
    if err != nil {
        log.Fatal("Unable to read input file " + filePath, err)
    }
    defer f.Close()

    csvReader := csv.NewReader(f)
    records, err := csvReader.ReadAll()
    if err != nil {
        log.Fatal("Unable to parse file as CSV for " + filePath, err)
    }

    return records
}

func startCountdown(seconds int, done chan<- struct{}) {
	fmt.Printf("Countdown started: %d seconds\n", seconds)
	time.Sleep(time.Duration(seconds) * time.Second)
	done <- struct{}{} // Send signal that the countdown is done
}

func interruptSignal() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
