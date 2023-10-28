package cmd

import (
	"fmt"
	"github.com/cslamar/audiobooker/audiobooker"
	"github.com/gen2brain/beeep"
	"time"
)

// cmdNotify output to the OS notification system
func cmdNotify(msg, title string) {
	if notify {
		beeep.Notify(fmt.Sprintf("Audiobooker | %s", title), msg, "")
	}
	if alert {
		beeep.Alert(fmt.Sprintf("Audiobooker | %s", title), msg, "")
	}
}

// notifyFinishedBook sends notification when book is complete
func notifyFinishedBook(book audiobooker.Book, startTime time.Time) {
	fmt.Println("Bind took:", time.Now().Sub(startTime).String())
	if notify {
		beeep.Notify("Audiobooker | Finished", fmt.Sprintf("%s - %s [%s]", book.Author, book.Title, time.Now().Sub(startTime).Round(1*time.Second)), "")
	}
	if alert {
		beeep.Alert("Audiobooker | Finished", fmt.Sprintf("%s - %s [%s]", book.Author, book.Title, time.Now().Sub(startTime).Round(1*time.Second)), "")
	}
}

// notifyError logs the error and sends notification
func notifyError(err error) {
	if notify {
		beeep.Notify("Audiobooker | Error", err.Error(), "")
	}
	if alert {
		beeep.Alert("Audiobooker | Error", err.Error(), "")
	}
}
