package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	locationCmd.PersistentFlags().StringVar(&config, "config", "", "")
	locationCmd.PersistentFlags().StringVar(&env, "environment", "", "")
	locationCmd.PersistentFlags().BoolVar(&test, "test", false, "")
	rootCmd.AddCommand(locationCmd)
}

var resources []ObjectID

var locationCmd = &cobra.Command{
	Use: "update-ead-locations",
	Run: func(cmd *cobra.Command, args []string) {
		setClient()
		updateLocations()
	},
}

func updateLocations() {
	getResourceIDs()
	resultChannel := make(chan []Result)
	resourceChunks := getChunks(resources)

	for i, resourceChunk := range resourceChunks {
		go updateLocation(resourceChunk, resultChannel, i+1)
	}

	t := time.Now()
	tf := t.Format("20060102-030405")
	var outfile *os.File
	if test {
		outfile, _ = os.Create("ead-locations-" + env + "-TEST-" + tf + ".tsv")

	} else {
		outfile, _ = os.Create("ead-locations-" + env + "-" + tf + ".tsv")
	}
	defer outfile.Close()
	writer := bufio.NewWriter(outfile)
	for range resourceChunks {
		for _, result := range <-resultChannel {
			writer.WriteString(result.String())
		}
		writer.Flush()
	}
}

func updateLocation(resourceChunk []ObjectID, resultChannel chan []Result, workerID int) {
	results := []Result{}
	fmt.Printf("* worker %d started, processing %d resources\n", workerID, len(resourceChunk))
	for i, resourceID := range resourceChunk {
		//log.Println("updating: ", resourceID.RepoID, resourceID.ObjectID)
		if i > 0 && i%500 == 0 {
			fmt.Printf("Worker %d has completed %d digital objects\n", workerID, i)
		}
		resource, err := client.GetResource(resourceID.RepoID, resourceID.ObjectID)
		if err != nil {
			results = append(results, Result{Code: "ERROR", URI: resourceID.String(), Msg: strings.ReplaceAll(err.Error(), "\n", ""), Time: time.Now(), Worker: workerID})
			continue
		}
		if !resource.Publish {
			results = append(results, Result{Code: "SKIPPED", URI: resource.URI, Msg: "Resource not set to Publish", Time: time.Now(), Worker: workerID})
			continue
		}
		faLocation := fmt.Sprintf("https://findingaids.library.nyu.edu/%s/%s", repositoryCodes[resourceID.RepoID], strings.ToLower(resource.MergeIDs("_")))
		jsonBytes := resource.Json
		resourceJson := JsonResponse{}
		if err := json.Unmarshal(jsonBytes, &resourceJson); err != nil {
			results = append(results, Result{Code: "ERROR", URI: resourceID.String(), Msg: strings.ReplaceAll(err.Error(), "\n", ""), Time: time.Now(), Worker: workerID})
			continue
		}

		resourceJson["ead_location"] = faLocation
		updateJson, err := json.Marshal(resourceJson)
		if err != nil {
			results = append(results, Result{Code: "ERROR", URI: resourceID.String(), Msg: strings.ReplaceAll(err.Error(), "\n", ""), Time: time.Now(), Worker: workerID})
			continue
		}

		if !test {
			code, msg, err := client.UpdateResourceJson(resourceID.RepoID, resourceID.ObjectID, updateJson)
			if err != nil {
				results = append(results, Result{Code: "ERROR", URI: resourceID.String(), Msg: strings.ReplaceAll(err.Error(), "\n", ""), Time: time.Now(), Worker: workerID})
				continue
			}

			result := Result{Code: "SUCCESS", URI: resource.URI, Msg: fmt.Sprintf("%d: %s", code, strings.ReplaceAll(msg, "\n", "")), Time: time.Now(), Worker: workerID}
			//log.Println(result)
			results = append(results, result)
		} else {
			results = append(results, Result{Code: "SKIPPED", URI: resourceID.String(), Msg: "TEST-MODE", Time: time.Now(), Worker: workerID})
			continue
		}
		if i > 1 && (i)%25 == 0 {
			fmt.Printf("* worker %d completed %d resources\n", workerID, i)
		}
	}

	resultChannel <- results
}
