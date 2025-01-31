package cmd

// implemented:
// $0 do refresh  --ao-uri|-a <ao URI> // updates the metadata of all DOs attached to the AO

// pending:
// $0 do create   --ao-uri|-a <ao URI> --file-uri|-f <file URI> --use-statement|-u <use statement>
// $0 do update   --ao-uri|-a <ao URI> --old-file-uri|-o <file URI to replace> --file-uri|-f <new file URI value> --use-statement|-u <new FV use statement>
import (
	"fmt"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

// --------------------------------------------------------------------------------
// Flags and Parameters
var aoURI string
var aoFlags = struct {
	URI      string
	URIShort string
}{
	URI:      "ao-uri",
	URIShort: "a",
}

var fileURI string // file uri
var fileURIFlags = struct {
	FileURI      string
	FileURIShort string
}{
	FileURI:      "file-uri",
	FileURIShort: "f",
}

var oldFileURI string // old file uri
var oldFileURIFlags = struct {
	OldFileURI      string
	OldFileURIShort string
}{
	OldFileURI:      "old-file-uri",
	OldFileURIShort: "o",
}

var useStatement string // use statement
var useStatementFlags = struct {
	UseStatement      string
	UseStatementShort string
}{
	UseStatement:      "use-statement",
	UseStatementShort: "u",
}

// --------------------------------------------------------------------------------
// doCmd represents the do command
var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Digital Object (do) operations",
	Long: `The 'do' noun allows you to perform
certain operations on Digital Object (do) resources`,
	Run: doRoot,
}

func doRoot(cmd *cobra.Command, args []string) { printNeedSubcommandHelp(cmd) }

// --------------------------------------------------------------------------------
var doRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh the titles of all Digital Objects",
	Long: `The refresh subcommand copies the title of the specified archival object (ao) 
to the title field of all digital objects ('do's) attached to the ao.`,
	RunE: doRefresh,
}

var doUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the File URI and/or Use Statement of a Digital Object File Version",
	Long: `The update subcommand allows one to change the File URI and/or the Use Statement
for a matching file version associated with an archival object (ao).  

The code retrieves the digital objects (dos) associated with the ao 
specified by the aoURI argument, and then updates any do file versions 
that match the --old-file-uri argument with the --file-uri argument 
and/or --use-statement argument.`,
	Args: doUpdateCheckArgs,
	RunE: doUpdate,
}

func init() {
	//----------------------------------------
	doCmd.PersistentFlags().StringVarP(&config, "config", "c", "", "path to the configuration file")
	doCmd.PersistentFlags().StringVarP(&env, "environment", "e", "", "environment to use")
	doCmd.PersistentFlags().BoolVar(&test, "test", false, "")

	//----------------------------------------
	doRefreshCmd.Flags().StringVarP(&aoURI, aoFlags.URI, aoFlags.URIShort, "", "uri of the archival object")
	doRefreshCmd.MarkFlagRequired(aoFlags.URI)

	//----------------------------------------
	doUpdateCmd.Flags().StringVarP(&aoURI, aoFlags.URI, aoFlags.URIShort, "", "uri of the archival object")
	doUpdateCmd.MarkFlagRequired(aoFlags.URI)

	doUpdateCmd.Flags().StringVarP(&oldFileURI, oldFileURIFlags.OldFileURI, oldFileURIFlags.OldFileURIShort, "", "file version URL to match [REQUIRED]")
	doUpdateCmd.MarkFlagRequired(oldFileURIFlags.OldFileURI)

	doUpdateCmd.Flags().StringVarP(&fileURI, fileURIFlags.FileURI, fileURIFlags.FileURIShort, "", "new file version URL [REQUIRED if --use-statement is not specified]")
	doUpdateCmd.Flags().StringVarP(&useStatement, useStatementFlags.UseStatement, useStatementFlags.UseStatementShort, "", "use statement for the new file version [REQUIRED if --file-uri is not specified]")

	// build command hierarchy
	rootCmd.AddCommand(doCmd)
	doCmd.AddCommand(doRefreshCmd)
	doCmd.AddCommand(doUpdateCmd)
}

func doRefresh(cmd *cobra.Command, args []string) (err error) {
	setClient()

	ao, err := client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		return err
	}

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		return err
	}

	// refresh the titles of all the DOs
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			return err
		}

		do.Title = ao.Title

		repoID, objectID, err := aspace.URISplit(doURI)
		if err != nil {
			return err
		}

		// update the do
		body, err := client.UpdateDigitalObject(repoID, objectID, do)
		if err != nil {
			return fmt.Errorf("%v : %s", err, body)
		}
		fmt.Printf("updated ao: %s do: %s\n", aoURI, doURI)
	}
	return nil
}

func doUpdateCheckArgs(cmd *cobra.Command, args []string) error {
	if fileURI == "" && useStatement == "" {
		return fmt.Errorf("--file-uri and/or --use-statement must be specified")
	}

	// arguments OK so disable cobra's usage output on error
	cmd.SilenceUsage = true

	return nil
}

func doUpdate(cmd *cobra.Command, args []string) (err error) {
	setClient()

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		return err
	}

	// determine whether we found a matching FileVersion.FileURI
	found := false

	// refresh the FileURI and UseStatement of all DOs with
	// a FileVersion.FileURImatching oldFileURI
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			return err
		}

		if do.FileVersions == nil {
			continue
		}

		// update the file version
		for i, fv := range do.FileVersions {
			if fv.FileURI == oldFileURI {
				found = true
				if fileURI != "" {
					do.FileVersions[i].FileURI = fileURI
				}
				if useStatement != "" {
					do.FileVersions[i].UseStatement = useStatement
				}
			}
		}

		// skip if we didn't find a matching FileURI
		if !found {
			continue
		}

		repoID, objectID, err := aspace.URISplit(doURI)
		if err != nil {
			return err
		}

		// update the do
		body, err := client.UpdateDigitalObject(repoID, objectID, do)
		if err != nil {
			return fmt.Errorf("%v : %s", err, body)
		}
		fmt.Printf("updated ao: %s do: %s\n", aoURI, doURI)
	}

	if !found {
		return fmt.Errorf("no matching file URI found for %s", oldFileURI)
	}
	return nil
}
