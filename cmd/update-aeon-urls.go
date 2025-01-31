package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	aeonCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	aeonCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "")
	aeonCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(aeonCmd)
}

var aeonCmd = &cobra.Command{
	Use: "update-aeon-urls",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()
		updateAeon()
	},
}

func updateAeon() {

	GetDOIDs()
	doChunks := getChunks(dos)

	resultChannel := make(chan []Result)

	for i, doChunk := range doChunks {
		go updateAeonLinks(doChunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102-030405")
	var outfile *os.File
	if test {
		outfile, _ = os.Create("update-aeon-urls-" + env + "-TEST-" + tf + ".tsv")

	} else {
		outfile, _ = os.Create("update-aeon-urls-" + env + "-" + tf + ".tsv")
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

func updateAeonLinks(doChunk []ObjectID, resultChannel chan []Result, worker int) {
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

		aeonLinks := false

		for _, fv := range do.FileVersions {
			if strings.Contains(fv.FileURI, "https://aeon.library.nyu.edu") {
				aeonLinks = true
				break
			}
		}

		if aeonLinks {
			newFV := updateAeonURI(do.FileVersions)
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
			results = append(results, Result{"SKIPPED", do.URI, "No aeon links in file versions", time.Now(), worker})
			continue
		}

	}

	fmt.Printf("Worker %d finished\n", worker)
	resultChannel <- results
}

func updateAeonURI(fvs []aspace.FileVersion) []aspace.FileVersion {
	newFileVersions := []aspace.FileVersion{}
	for _, fv := range fvs {
		if strings.Contains(fv.FileURI, "https://aeon.library.nyu.edu") {
			fv.FileURI = "https://hdl.handle.net/2333.1/material-request-placeholder"
		}
		newFileVersions = append(newFileVersions, fv)
	}
	return newFileVersions
}
