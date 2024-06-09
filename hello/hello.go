package main

import (
    "fmt"
    "log"
		"rsc.io/quote"
    "example.com/greetings"
)

func main() {
  // Set properties of the predefined Logger, including
    // the log entry prefix and a flag to disable printing
    // the time, source file, and line number.
    log.SetPrefix("greetings: ")
    log.SetFlags(0)
   
    message, err := greetings.Hello("")
    // message := greetings.Hello("Tia")
    if err != nil {
      log.Fatal(err)
  }
    fmt.Println(message)
		fmt.Println(quote.Go())
}