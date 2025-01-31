package cmd

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	convertHandlesCmd.PersistentFlags().StringVar(&config, "config", "", "")
	convertHandlesCmd.PersistentFlags().StringVar(&env, "environment", "", "")
	convertHandlesCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(convertHandlesCmd)
}

var httpMatch = regexp.MustCompile("^http://hdl.handle")

var convertHandlesCmd = &cobra.Command{
	Use: "convert-http-handles",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()
		convertHandles()
		if err != nil {
			panic(err)
		}
	},
}

func convertHandles() {
	GetDOIDs()
	doChunks := getChunks(dos)
	resultChannel := make(chan []Result)
	for i, doChunk := range doChunks {
		go updateHttp(doChunk, resultChannel, i+1)
	}
	t := time.Now()
	tf := t.Format("20060102-030405")
	var outfile *os.File
	if test {
		outfile, _ = os.Create("convert-http-handles-" + env + "-TEST-" + tf + ".tsv")

	} else {
		outfile, _ = os.Create("convert-http-handles-" + env + "-" + tf + ".tsv")
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	writer.WriteString("timestamp\tworkerID\tresult\taspaceURL\tmessage\n")
	writer.Flush()

	for range doChunks {
		results := <-resultChannel
		for _, result := range results {
			writer.WriteString(fmt.Sprintf("%s\t%d\t%s\t%s\t%s\n", result.Time.Format(time.RFC3339), result.Worker, result.Code, result.URI, result.Msg))
			writer.Flush()
		}
	}
	writer.Flush()
}

func updateHttp(doChunk []ObjectID, resultChannel chan []Result, worker int) {
	results := []Result{}
	fmt.Printf("Worker %d started, processing %d records\n", worker, len(doChunk))

	for i, doid := range doChunk {
		if i > 0 && i%500 == 0 {
			fmt.Printf("Worker %d has completed %d digital objects\n", worker, i)
		}

		//request the digital object
		do, err := client.GetDigitalObject(doid.RepoID, doid.ObjectID)
		if err != nil {
			results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now(), worker})
			continue
		}

		http := false
		for _, fv := range do.FileVersions {
			if httpMatch.MatchString(fv.FileURI) {
				http = true
				break
			}
		}

		if http {
			newFV := updateHttpURI(do.FileVersions)
			do.FileVersions = newFV
			if !test {
				response, err := client.UpdateDigitalObject(doid.RepoID, doid.ObjectID, do)
				if err != nil {
					results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now(), worker})
					continue
				}
				results = append(results, Result{"UPDATED", do.URI, strings.ReplaceAll(response, "\n", ""), time.Now(), worker})
				continue
			} else {
				results = append(results, Result{"SKIPPED", do.URI, "Test-Mode, DO update skipped", time.Now(), worker})
				continue
			}
		} else {
			results = append(results, Result{"SKIPPED", do.URI, "No http handles found in file versions", time.Now(), worker})
			continue
		}
	}

	fmt.Printf("Worker %d finished\n", worker)
	resultChannel <- results
}

func updateHttpURI(fvs []aspace.FileVersion) []aspace.FileVersion {
	newFileVersions := []aspace.FileVersion{}
	for _, fv := range fvs {
		if httpMatch.MatchString(fv.FileURI) {
			fv.FileURI = strings.ReplaceAll(fv.FileURI, "http://hdl.handle", "https://hdl.handle")
		}
		newFileVersions = append(newFileVersions, fv)
	}
	return newFileVersions
}
