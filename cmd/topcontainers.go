package cmd

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"regexp"
	"strings"
	"time"
)

func init() {
	tcCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	rootCmd.AddCommand(tcCmd)
}

var (
	tcids    = []ObjectID{}
	tcChunks = [][]ObjectID{}
	erecs    = regexp.MustCompile("electronic records")
)

var tcCmd = &cobra.Command{
	Use: "top-containers",
	Run: func(cmd *cobra.Command, args []string) {
		getClient()
		getTCids()
		tcChunks = getChunks(tcids)

		resultChannel := make(chan []Result)

		for i, tcChunk := range tcChunks {
			go removeTopContainers(tcChunk, resultChannel, i+1)
		}

		t := time.Now()
		tf := t.Format("20060102T15:04")
		outfile, _ := os.Create("topcontainers-" + tf + ".tsv")
		defer outfile.Close()
		writer := bufio.NewWriter(outfile)

		for range tcChunks {
			for _, result := range <-resultChannel {
				writer.WriteString(result.String())
			}
			writer.Flush()
		}
	},
}

func removeTopContainers(tcChunk []ObjectID, resultChannel chan []Result, worker int) {
	results := []Result{}
	fmt.Printf("Worker %d processing %d top containers\n", worker, len(tcChunk))
	for i, objectId := range tcChunk {
		if i-1 > 0 && (i-1)%100 == 0 {
			fmt.Printf("Worker %d completed %d top containers\n", worker, i-1)
		}
		tc, err := client.GetTopContainer(objectId.RepoID, objectId.ObjectID)
		if err != nil {
			results = append(results, Result{"ERROR", "", err.Error(), time.Now(), worker})
			continue
		}
		if erecs.MatchString(strings.ToLower(tc.DisplayString)) {
			msg, err := client.DeleteTopContainer(objectId.RepoID, objectId.ObjectID)
			if err != nil {
				results = append(results, Result{"ERROR", tc.URI, err.Error(), time.Now(), worker})
				continue
			}
			results = append(results, Result{"DELETED", tc.URI, strings.ReplaceAll(msg, "\n", ""), time.Now(), worker})
			continue
		}
		results = append(results, Result{"SKIPPED", tc.URI, "", time.Now(), worker})
	}
	resultChannel <- results
}

func getTCids() {
	for _, rid := range []int{2, 3, 6} {
		ids, err := client.GetTopContainerIDs(rid)
		if err != nil {
			panic(err)
		}
		for _, id := range ids {
			tcids = append(tcids, ObjectID{rid, id})
		}
	}
}
