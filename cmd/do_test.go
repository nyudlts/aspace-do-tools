package cmd

import (
	"fmt"
	"testing"

	"github.com/nyudlts/go-aspace"
)

func TestDORefresh(t *testing.T) {
	var args []string
	aoURI := "/repositories/3/archival_objects/512700"

	setClient()

	ao, err := client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("before test: could not get Archival Object")
		t.FailNow()
	}

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("before test: could not get Digital Object URIs")
		t.FailNow()
	}

	want := ""
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			t.Errorf("before test: could not get Digital Object %s", doURI)
		}

		if do.Title != ao.Title {
			t.Errorf("before test: Digital Object title does not match Archival Object title: %s != %s", do.Title, ao.Title)
			t.FailNow()
		}
		want += fmt.Sprintf("updated ao: %s do: %s\n", aoURI, doURI)
	}

	saveAOTitle := ao.Title

	// change the title of the Archival Object
	ao.Title = "waffles are delicious"
	repoID, aoObjectID, err := aspace.URISplit(aoURI)
	if err != nil {
		t.Errorf("before test: could not split AO URI: %s", aoURI)
		t.FailNow()
	}

	body, err := client.UpdateArchivalObject(repoID, aoObjectID, &ao)
	if err != nil {
		t.Errorf("%v, %s", err, body)
	}

	// RUN THE COMMAND
	// refresh the Digital Objects
	setCmdFlag(doRefreshCmd, aoFlags.URI, aoURI)
	got, errOut, err := CaptureCmdStdoutStderrE(doRefresh, doRefreshCmd, args)
	if err != nil {
		t.Errorf("doRefresh: %v, %v", err, errOut)
	}

	if want != got {
		t.Errorf("wanted: %s\n got: %s\n", want, got)
	}

	// reset the title of the Archival Object
	// need to fetch the AO again or else we get a stale object error
	ao, err = client.GetArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("post test: could not get Archival Object")
	}

	ao.Title = saveAOTitle
	body, err = client.UpdateArchivalObject(repoID, aoObjectID, &ao)
	if err != nil {
		t.Errorf("during AO restore: %v, %s", err, body)
	}

	// reset the title of the Digital Object(s)
	for _, doURI := range doURIs {
		do, err := client.GetDigitalObjectFromURI(doURI)
		if err != nil {
			t.Errorf("post test: could not get Digital Object %s", doURI)
		}

		do.Title = ao.Title
		repoID, doObjectID, err := aspace.URISplit(doURI)
		if err != nil {
			t.Errorf("post test: could not split DO URI: %s", doURI)
			t.FailNow()
		}

		body, err := client.UpdateDigitalObject(repoID, doObjectID, &do)

		if err != nil {
			t.Errorf("post test: %v : %s", err, body)
		}
	}
}

func TestDOUpdate(t *testing.T) {

	scenarios := [][]string{
		// "aoURI", "oldFileURI", "oldUseStatement", "newFileURI", "newUseStatement"
		{"/repositories/3/archival_objects/512700", "https://hdl.handle.net/2333.1/djh9w43b", "image-service", "https://hdl.handle.net/2333.1/material-request-placeholder", "audio-reading-room"},
		{"/repositories/3/archival_objects/512700", "https://hdl.handle.net/2333.1/djh9w43b", "image-service", "https://hdl.handle.net/2333.1/bananas-are-great", ""},
		{"/repositories/3/archival_objects/512700", "https://hdl.handle.net/2333.1/djh9w43b", "image-service", "", "video-reading-room"},
		{"/repositories/3/archival_objects/512700", "https://hdl.handle.net/2333.1/djh9w43b", "image-service", "", ""},
	}

	for s := range scenarios {
		runDOUPdateTest(t, scenarios[s][0], scenarios[s][1], scenarios[s][2], scenarios[s][3], scenarios[s][4])
	}

}

func runDOUPdateTest(t *testing.T, aoURI, oldFileURI, oldUseStatement, newFileURI, newUseStatement string) {
	var args []string

	setClient()

	doURIs, err := client.GetDigitalObjectIDsForArchivalObjectFromURI(aoURI)
	if err != nil {
		t.Errorf("before test: could not get Digital Object URIs")
		t.FailNow()
	}

	if len(doURIs) != 1 {
		t.Errorf("before test: unexpected number of Digital Object URIs: %d", len(doURIs))
		t.FailNow()
	}

	want := ""
	doURI := doURIs[0]
	do, err := client.GetDigitalObjectFromURI(doURI)
	if err != nil {
		t.Errorf("before test: could not get Digital Object %s", doURI)
	}

	if len(do.FileVersions) != 1 {
		t.Errorf("before test: Digital Object has unexpected number of FileVersions: %d", len(do.FileVersions))
		t.FailNow()
	}

	fv := do.FileVersions[0]
	if fv.FileURI != oldFileURI {
		t.Errorf("before test: Digital Object file version has unexpected FileURI: %s != %s", fv.FileURI, oldFileURI)
		t.FailNow()
	}
	if fv.UseStatement != oldUseStatement {
		t.Errorf("before test: Digital Object file version has unexpected UseStatement: %s != %s", fv.UseStatement, oldUseStatement)
		t.FailNow()
	}
	want += fmt.Sprintf("updated ao: %s do: %s\n", aoURI, doURI)

	// RUN THE COMMAND
	// refresh the Digital Objects
	setCmdFlag(doUpdateCmd, aoFlags.URI, aoURI)
	setCmdFlag(doUpdateCmd, oldFileURIFlags.OldFileURI, oldFileURI)
	setCmdFlag(doUpdateCmd, fileURIFlags.FileURI, newFileURI)
	setCmdFlag(doUpdateCmd, useStatementFlags.UseStatement, newUseStatement)

	got, errOut, err := CaptureCmdStdoutStderrE(doUpdate, doUpdateCmd, args)
	if err != nil {
		t.Errorf("doUpdate: %v, %v", err, errOut)
	}

	if want != got {
		t.Errorf("wanted: %s\n got: %s\n", want, got)
	}

	// reset the FileURI and UseStatement of the DO
	// need to fetch the DO again or else we get a stale object error
	do, err = client.GetDigitalObjectFromURI(doURI)
	if err != nil {
		t.Errorf("post test: could not get Digital Object")
	}

	// assert that the values that were supposed to be changed were actually changed
	if newFileURI == "" {
		// nothing should have changed
		if do.FileVersions[0].FileURI != oldFileURI {
			t.Errorf("post test: Digital Object file version has unexpected FileURI: %s != %s", do.FileVersions[0].FileURI, newFileURI)
		}
	} else {
		// the fileURI should have changed
		if do.FileVersions[0].FileURI != newFileURI {
			t.Errorf("post test: Digital Object file version has unexpected FileURI: %s != %s", do.FileVersions[0].FileURI, newFileURI)
		}
	}

	if newUseStatement == "" {
		// nothing should have changed
		if do.FileVersions[0].UseStatement != oldUseStatement {
			t.Errorf("post test: Digital Object file version has unexpected UseStatement: %s != %s", do.FileVersions[0].UseStatement, oldUseStatement)
		}
	} else {
		// the UseStatement should have changed
		if do.FileVersions[0].UseStatement != newUseStatement {
			t.Errorf("post test: Digital Object file version has unexpected UseStatement: %s != %s", do.FileVersions[0].UseStatement, newUseStatement)
		}
	}

	// reset the Digital Object(s)
	do.FileVersions[0].FileURI = oldFileURI
	do.FileVersions[0].UseStatement = oldUseStatement

	repoID, doObjectID, err := aspace.URISplit(doURI)
	if err != nil {
		t.Errorf("post test: could not split DO URI: %s", doURI)
		t.FailNow()
	}

	body, err := client.UpdateDigitalObject(repoID, doObjectID, &do)

	if err != nil {
		t.Errorf("post test: %v : %s", err, body)
	}
}
