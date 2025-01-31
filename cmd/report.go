package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

func init() {
	reportCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "")
	reportCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "")
	rootCmd.AddCommand(reportCmd)
}

var outliers = []string{"audio-master", "video-master", "image-master", "image-master-edited", "electronic-records-master", "undefined"}

func isOutlier(role string) bool {
	for _, r := range outliers {
		if r == role {
			fmt.Println(true, role)
			return true
		}
	}
	return false
}

//var outfile *os.File

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report of use statements",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running Role Report")
		setClient()
		ReportDOs()
	},
}

func ReportDOs() {

	GetDOIDs()
	doChunks := getChunks(dos)

	resultsChannel := make(chan map[string]int)

	//get the dos
	undefUrl, _ := os.Create("service-roles.tsv")
	udefWriter := bufio.NewWriter(undefUrl)

	for i, chunk := range doChunks {
		go GetRoles(chunk, resultsChannel, i+1, udefWriter)
	}

	results := map[string]int{}
	for range doChunks {
		chunk := <-resultsChannel
		for k, v := range chunk {
			if HasRole(results, k) {
				results[k] = results[k] + v
			} else {
				results[k] = v
			}
		}
	}
	GenerateRoleReport(results)
	PrintRoleMap(results)
	udefWriter.Flush()
}

func GenerateRoleReport(roles map[string]int) {
	t := time.Now()
	tf := t.Format("20060102-15-04")
	outfile, err := os.Create("roles-report-" + tf + ".tsv")
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	writer := bufio.NewWriter(outfile)
	for k, v := range roles {
		writer.WriteString(fmt.Sprintf("%s\t%d\n", k, v))
	}
	writer.Flush()
}

func HasRole(roles map[string]int, role string) bool {
	for k := range roles {
		if k == role {
			return true
		}
	}
	return false
}

func GetRoles(chunk []ObjectID, resultsChannel chan map[string]int, worker int, writer *bufio.Writer) {
	results := map[string]int{}
	fmt.Printf("Worker %d processing %d records\n", worker, len(chunk))
	for i, doid := range chunk {
		if i > 0 && i%100 == 0 {
			fmt.Printf("Worker %d has completed %d records\n", worker, i)
		}
		do, err := client.GetDigitalObject(doid.RepoID, doid.ObjectID)
		if err != nil {
			log.Println(err.Error())
		}
		fileVersions := do.FileVersions
		if len(fileVersions) > 0 {
			for _, fileVersion := range fileVersions {
				role := fileVersion.UseStatement
				if role == "" {
					role = "undefined"
				}

				if isOutlier(role) && do.Publish {
					writer.WriteString(fmt.Sprintf("%s\t%s\n", role, do.URI))
					writer.Flush()
				}

				if HasRole(results, role) {
					results[role] = results[role] + 1
				} else {
					results[role] = 1
				}
			}

		}
	}
	resultsChannel <- results
}

func PrintRoleMap(roles map[string]int) {
	roleKeys := []string{}
	for k := range roles {
		roleKeys = append(roleKeys, k)
	}
	sort.Strings(roleKeys)
	for _, k := range roleKeys {
		fmt.Printf("%s\t%d\n", k, roles[k])
	}
}
