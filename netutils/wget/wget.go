package wget

import (
	"fmt"
	"github.com/cavaliercoder/grab"
	"os"
	"time"
)

const Name = "wget"

type cmd struct{}

func New() cmd { return cmd{} }

func (cmd) String() string { return Name }
func (cmd) Usage() string  { return Name + ` URL...` }

func (cmd) Main(args ...string) error {
	// validate command args
	if len(args) < 1 {
		return fmt.Errorf("URL: missing")
	}

	// create a custom client
	client := grab.NewClient()
	client.UserAgent = "Platina Go-ES"

	// create request for each URL given on the command line
	reqs := make([]*grab.Request, 0)
	for _, url := range args {
		req, err := grab.NewRequest(url)
		if err != nil {
			return err
		}

		reqs = append(reqs, req)
	}

	// start file downloads, 3 at a time
	fmt.Printf("Downloading %d files...\n", len(reqs))
	respch := client.DoBatch(3, reqs...)

	// start a ticker to update progress every 200ms
	t := time.NewTicker(200 * time.Millisecond)

	// monitor downloads
	completed := 0
	inProgress := 0
	responses := make([]*grab.Response, 0)
	for completed < len(reqs) {
		select {
		case resp := <-respch:
			// a new response has been received and has started downloading
			// (nil is received once, when the channel is closed by grab)
			if resp != nil {
				responses = append(responses, resp)
			}

		case <-t.C:
			// clear lines
			if inProgress > 0 {
				fmt.Printf("\033[%dA\033[K", inProgress)
			}

			// update completed downloads
			for i, resp := range responses {
				if resp != nil && resp.IsComplete() {
					// print final result
					if resp.Error != nil {
						fmt.Fprintf(os.Stderr, "Error downloading %s: %v\n", resp.Request.URL(), resp.Error)
					} else {
						fmt.Printf("Finished %s %d / %d bytes (%d%%)\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
					}

					// mark completed
					responses[i] = nil
					completed++
				}
			}

			// update downloads in progress
			inProgress = 0
			for _, resp := range responses {
				if resp != nil {
					inProgress++
					fmt.Printf("Downloading %s %d / %d bytes (%d%%)\033[K\n", resp.Filename, resp.BytesTransferred(), resp.Size, int(100*resp.Progress()))
				}
			}
		}
	}

	t.Stop()

	fmt.Printf("%d files successfully downloaded.\n", len(reqs))
	return nil
}
