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
	thumbnailCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	thumbnailCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "")
	thumbnailCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(thumbnailCmd)
}

var thumbnailCmd = &cobra.Command{
	Use: "remove-thumbnails",
	Run: func(cmd *cobra.Command, args []string) {
		client, err = aspace.NewClient(config, env, 20)
		if err != nil {
			panic(err)
		}
		GetDOIDs()
		removeThumbs()
	},
}

func removeThumbs() {

	chunks := getChunks(dos)
	resultChannel := make(chan []Result)

	for i, chunk := range chunks {
		go removeThumbnails(chunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102-030405")
	var outfile *os.File
	if test {
		outfile, _ = os.Create("remove-thumbnails-" + env + "-TEST-" + tf + ".tsv")

	} else {
		outfile, _ = os.Create("remove-thumbnails-" + env + "-" + tf + ".tsv")
	}

	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	writer.WriteString("timestamp\tworkerID\tresult\taspaceURL\tmessage\n")
	writer.Flush()

	for range chunks {
		results := <-resultChannel
		for _, result := range results {
			writer.WriteString(fmt.Sprintf("%s\t%d\t%s\t%s\t%s\n", result.Time.Format(time.RFC3339), result.Worker, result.Code, result.URI, result.Msg))
			writer.Flush()
		}
	}

}

func removeThumbnails(chunk []ObjectID, resultChannel chan []Result, worker int) {
	results := []Result{}
	fmt.Printf("Worker %d started, processing %d records\n", worker, len(chunk))
	for i, doid := range chunk {
		if i > 0 && i%250 == 0 {
			fmt.Printf("Worker %d has completed %d digital objects\n", worker, i)
		}

		//request the digital object
		do, err := client.GetDigitalObject(doid.RepoID, doid.ObjectID)
		if err != nil {
			results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now(), worker})
			continue
		}

		//check that there are file versions
		if len(do.FileVersions) > 0 {

			//check for thumnbnails
			if do.ContainsUseStatement("image-thumbnail") {
				//delete any dos that only have a thumbnail
				if len(do.FileVersions) == 1 {
					if !test {
						response, err := client.DeleteDigitalObject(doid.RepoID, doid.ObjectID)
						if err != nil {
							results = append(results, Result{"ERROR", do.URI, err.Error(), time.Now(), worker})
							continue
						}
						results = append(results, Result{"DELETED", do.URI, strings.ReplaceAll(response, "\n", ""), time.Now(), worker})
						continue
					} else {
						results = append(results, Result{"SKIPPED", do.URI, "TEST-MODE, DO deletion skipped", time.Now(), worker})
						continue
					}
				}

				//update dos with more than one file versions
				do.FileVersions = updateFileVersions(do)
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
				results = append(results, Result{"SKIPPED", do.URI, "No image-thumbnails in file versions", time.Now(), worker})
				continue
			}
		}
		results = append(results, Result{"SKIPPED", do.URI, "No file versions", time.Now(), worker})
	}
	fmt.Printf("Worker %d finished\n", worker)
	resultChannel <- results
}

func updateFileVersions(do aspace.DigitalObject) []aspace.FileVersion {
	newFileVersions := []aspace.FileVersion{}
	for _, fileVersion := range do.FileVersions {
		if fileVersion.UseStatement != "image-thumbnail" {
			newFileVersions = append(newFileVersions, fileVersion)
		}
	}
	return newFileVersions
}
