package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/nyudlts/go-aspace"
)

var (
	client          *aspace.ASClient
	env             string
	dos             = []ObjectID{}
	workers         = 12
	err             error
	config          string
	repositoryCodes      = map[int]string{2: "tamwag", 3: "fales", 6: "nyuarchives"}
	test            bool = true
)

type JsonResponse map[string]interface{}

type ObjectID struct {
	RepoID   int
	ObjectID int
}

func (o ObjectID) String() string {
	return fmt.Sprintf("Repository ID: %d, Object ID: %d", o.RepoID, o.ObjectID)
}

type Result struct {
	Code   string
	URI    string
	Msg    string
	Time   time.Time
	Worker int
}

func setClient() {
	client, err = aspace.NewClient(config, env, 20)
	if err != nil {
		panic(err)
	}
}

func (r Result) String() string {
	return fmt.Sprintf("%s\t%d\t%s\t%s\t%s\n", r.Time.Format(time.RFC3339), r.Worker, r.Code, r.URI, r.Msg)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getChunks(oids []ObjectID) [][]ObjectID {
	var divided [][]ObjectID

	chunkSize := (len(oids) + workers - 1) / workers

	for i := 0; i < len(oids); i += chunkSize {
		end := i + chunkSize

		if end > len(oids) {
			end = len(oids)
		}

		divided = append(divided, oids[i:end])
	}
	return divided
}

func GetDOIDs() {
	for _, repoId := range []int{2, 3, 6} {
		doIDs, err := client.GetDigitalObjectIDs(repoId)
		if err != nil {
			panic(err)
		}
		for _, digitalObjectID := range doIDs {
			dos = append(dos, ObjectID{repoId, digitalObjectID})
		}
	}
}

func getResourceIDs() {
	resources = []ObjectID{}
	for _, repID := range []int{2, 3, 6} {
		resourceIDs, err := client.GetResourceIDs(repID)
		if err != nil {
			panic(err)
		}
		for _, resourceID := range resourceIDs {
			resources = append(resources, ObjectID{repID, resourceID})
		}
	}
}
